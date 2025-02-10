package s2

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/jakthom/s3c/pkg/middleware"
	zlog "github.com/rs/zerolog/log"
)

var (
	// bucketNameValidator is a regex for validating bucket names
	bucketNameValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_\.]{1,255}/`)
	// authV4HeaderValidator is a regex for validating the authorization
	// header when using AWs' auth V4
	authV4HeaderValidator = regexp.MustCompile(`^AWS4-HMAC-SHA256 Credential=([^/]*)/([^/]*)/([^/]*)/s3/aws4_request, ?SignedHeaders=([^,]+), ?Signature=(.+)$`)

	// subresourceQueryParams is a list of query parameters that are
	// considered queries for "subresources" in S3. This is used in
	// auth validation.
	subresourceQueryParams = []string{
		"acl",
		"lifecycle",
		"location",
		"logging",
		"notification",
		"partNumber",
		"policy",
		"requestPayment",
		"torrent",
		"uploadId",
		"uploads",
		"versionId",
		"versioning",
		"versions",
	}
)

// attachBucketRoutes adds bucket-related routes to a router
func attachBucketRoutes(router *mux.Router, handler *bucketHandler, multipartHandler *multipartHandler, objectHandler *objectHandler) {
	router.Methods("GET").Queries("versioning", "").HandlerFunc(handler.versioning)
	router.Methods("PUT").Queries("versioning", "").HandlerFunc(handler.setVersioning)
	router.Methods("GET").Queries("versions", "").HandlerFunc(handler.listVersions)
	router.Methods("GET").Queries("uploads", "").HandlerFunc(multipartHandler.list)
	router.Methods("GET").Queries("location", "").HandlerFunc(handler.location)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get) // TODO -> implement HEAD handler
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("POST").Queries("delete", "").HandlerFunc(objectHandler.post)
	router.Methods("DELETE").HandlerFunc(handler.del)

}

// attachBucketRoutes adds object-related routes to a router
func attachObjectRoutes(router *mux.Router, handler *objectHandler, multipartHandler *multipartHandler) {

	router.Methods("GET").Queries("uploadId", "").HandlerFunc(multipartHandler.listChunks)
	router.Methods("POST").Queries("uploads", "").HandlerFunc(multipartHandler.init)
	router.Methods("POST").Queries("uploadId", "").HandlerFunc(multipartHandler.complete)
	router.Methods("PUT").Queries("uploadId", "").HandlerFunc(multipartHandler.put)
	router.Methods("DELETE").Queries("uploadId", "").HandlerFunc(multipartHandler.del)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get) // TODO -> implement HEAD handler
	router.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(handler.copy)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

// S2 is the root struct used in the s2 library
type S2 struct {
	Auth                 AuthController
	Service              ServiceController
	Bucket               BucketController
	Object               ObjectController
	Multipart            MultipartController
	maxRequestBodyLength uint32
	readBodyTimeout      time.Duration
}

// NewS2 creates a new S2 instance. One created, you set zero or more
// attributes to implement various S3 functionality, then create a router.
// `maxRequestBodyLength` specifies maximum request body size; if the value is
// 0, there is no limit. `readBodyTimeout` specifies the maximum amount of
// time s2 should spend trying to read the body of requests.
func NewS2(maxRequestBodyLength uint32, readBodyTimeout time.Duration) *S2 {
	return &S2{
		Auth:                 nil,
		Service:              unimplementedServiceController{},
		Bucket:               unimplementedBucketController{},
		Object:               unimplementedObjectController{},
		Multipart:            unimplementedMultipartController{},
		maxRequestBodyLength: maxRequestBodyLength,
		readBodyTimeout:      readBodyTimeout,
	}
}

// authV4 validates a request using AWS' auth V4
func (h *S2) authV4(w http.ResponseWriter, r *http.Request, auth string) error { // TODO -> break this functionality out
	// parse auth-related headers
	match := authV4HeaderValidator.FindStringSubmatch(auth)
	if len(match) == 0 {
		return AuthorizationHeaderMalformedError(r)
	}

	accessKey := match[1]
	date := match[2]
	region := match[3]
	signedHeaderKeys := strings.Split(match[4], ";")
	sort.Strings(signedHeaderKeys)
	expectedSignature := match[5]

	// get the expected secret key
	secretKey, err := h.Auth.SecretKey(r, accessKey, &region)
	if err != nil {
		return InternalError(r, err)
	}
	if secretKey == nil {
		return InvalidAccessKeyIDError(r)
	}

	// step 1: construct the canonical request
	var signedHeaders strings.Builder
	for _, key := range signedHeaderKeys {
		signedHeaders.WriteString(key)
		signedHeaders.WriteString(":")
		if key == "host" {
			signedHeaders.WriteString(r.Host)
		} else {
			signedHeaders.WriteString(strings.TrimSpace(r.Header.Get(key)))
		}
		signedHeaders.WriteString("\n")
	}

	canonicalRequest := strings.Join([]string{
		r.Method,
		normURI(r.URL.Path),
		normQuery(r.URL.Query()),
		signedHeaders.String(),
		strings.Join(signedHeaderKeys, ";"),
		r.Header.Get("x-amz-content-sha256"),
	}, "\n")

	timestamp, err := parseAWSTimestamp(r)
	if err != nil {
		return err
	}
	formattedTimestamp := formatAWSTimestamp(timestamp)

	// step 2: construct the string to sign
	stringToSign := fmt.Sprintf(
		"AWS4-HMAC-SHA256\n%s\n%s/%s/s3/aws4_request\n%x",
		formattedTimestamp,
		date,
		region,
		sha256.Sum256([]byte(canonicalRequest)),
	)

	// step 3: calculate the signing key
	dateKey := hmacSHA256([]byte("AWS4"+*secretKey), date)
	dateRegionKey := hmacSHA256(dateKey, region)
	dateRegionServiceKey := hmacSHA256(dateRegionKey, "s3")
	signingKey := hmacSHA256(dateRegionServiceKey, "aws4_request")

	// step 4: construct & verify the signature
	signature := hmacSHA256(signingKey, stringToSign)

	if expectedSignature != fmt.Sprintf("%x", signature) {
		return SignatureDoesNotMatchError(r)
	}

	vars := mux.Vars(r)
	vars["authMethod"] = "v4"
	vars["authAccessKey"] = accessKey
	vars["authRegion"] = region
	// store signature data as vars, since it may be reused for verifying chunked uploads
	vars["authSignature"] = expectedSignature
	// This is a bit unfortunate -- `vars` can only store string values, so we need to
	// convert the bytes to a string. Note that this string may not be valid,
	// i.e. it may contain non-utf8 sequences.
	vars["authSignatureKey"] = string(signingKey)
	vars["authSignatureTimestamp"] = formattedTimestamp
	vars["authSignatureDate"] = date
	vars["authSignatureRegion"] = region
	return nil
}

// authMiddleware creates a middleware handler for dealing with AWS auth
func (h *S2) authMiddleware(next http.Handler) http.Handler { // TODO -> add more flexible middleware
	// Verifies auth using AWS v4 auth mechanisms. Much of the code is
	// built off of smartystreets/go-aws-auth, which does signing from the
	// client-side:
	// https://github.com/smartystreets/go-aws-auth
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("authorization")

		passed := true
		var err error
		if strings.HasPrefix(auth, "AWS4-HMAC-SHA256 ") {
			err = h.authV4(w, r, auth)
		} else {
			passed, err = h.Auth.CustomAuth(r)
			vars := mux.Vars(r)
			vars["authMethod"] = "custom"
		}
		if err != nil {
			WriteError(w, r, err)
			return
		}
		if !passed {
			WriteError(w, r, AccessDeniedError(r))
			return
		}

		next.ServeHTTP(w, r)
	})
}

// bodyReadingMiddleware creates a middleware for reading request bodies
func (h *S2) bodyReadingMiddleware(next http.Handler) http.Handler { // TODO -> break out middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentLengthStr, ok := singleHeader(r, "Content-Length")
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		contentLength, err := strconv.ParseUint(contentLengthStr, 10, 32)
		if err != nil {
			WriteError(w, r, InvalidArgumentError(r))
			return
		}
		if h.maxRequestBodyLength > 0 && uint32(contentLength) > h.maxRequestBodyLength {
			WriteError(w, r, EntityTooLargeError(r))
			return
		}

		body := []byte{}

		if contentLength > 0 {
			bodyBuf, err := h.readBody(r, uint32(contentLength))
			if err != nil {
				WriteError(w, r, err)
				return
			}
			if bodyBuf == nil {
				WriteError(w, r, RequestTimeoutError(r))
				return
			}
			body = bodyBuf.Bytes()
			r.Body = io.NopCloser(bodyBuf)
		} else {
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		expectedSHA256, ok := singleHeader(r, "x-amz-content-sha256")
		if ok {
			if len(expectedSHA256) != 64 {
				WriteError(w, r, InvalidDigestError(r))
				return
			}
			actualSHA256 := sha256.Sum256(body)
			if fmt.Sprintf("%x", actualSHA256) != expectedSHA256 {
				WriteError(w, r, BadDigestError(r))
				return
			}
		}

		expectedMD5, ok := singleHeader(r, "Content-Md5")
		if ok {
			expectedMD5Decoded, err := base64.StdEncoding.DecodeString(expectedMD5)
			if err != nil || len(expectedMD5Decoded) != 16 {
				WriteError(w, r, InvalidDigestError(r))
				return
			}
			actualMD5 := md5.Sum(body)
			if !bytes.Equal(expectedMD5Decoded, actualMD5[:]) {
				WriteError(w, r, BadDigestError(r))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// readBody efficiently reads a request body, or times out
func (h *S2) readBody(r *http.Request, length uint32) (*bytes.Buffer, error) {
	var body bytes.Buffer
	body.Grow(int(length))

	ch := make(chan error)
	go func() {
		n, err := body.ReadFrom(r.Body)
		r.Body.Close()
		if err != nil {
			ch <- err
		}
		if uint32(n) != length {
			ch <- IncompleteBodyError(r)
		}
		ch <- nil
	}()

	select {
	case err := <-ch:
		if err != nil {
			return nil, err
		}
		return &body, nil
	case <-time.After(h.readBodyTimeout):
		return nil, nil
	}
}

// Router creates a new mux router.
func (h *S2) Router() *mux.Router {
	serviceHandler := &serviceHandler{
		controller: h.Service,
	}
	bucketHandler := &bucketHandler{
		controller: h.Bucket,
	}
	objectHandler := &objectHandler{
		controller: h.Object,
	}
	multipartHandler := &multipartHandler{
		controller: h.Multipart,
	}

	router := mux.NewRouter()
	router.Use(middleware.RequestIdMiddleware)
	if h.Auth != nil {
		router.Use(h.authMiddleware)
	}
	router.Use(middleware.EtagMiddleware)
	router.Use(h.bodyReadingMiddleware)

	router.Path(`/`).Methods("GET", "HEAD").HandlerFunc(serviceHandler.get)

	// Bucket-related routes. Repo validation regex is the same that the aws
	// cli uses. There's two routers - one with a trailing a slash and one
	// without. Both route to the same handlers, i.e. a request to `/foo` is
	// the same as `/foo/`. This is used instead of mux's builtin "strict
	// slash" functionality, because that uses redirects which doesn't always
	// play nice with s3 clients.
	trailingSlashBucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/`).Subrouter()
	attachBucketRoutes(trailingSlashBucketRouter, bucketHandler, multipartHandler, objectHandler)
	bucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}`).Subrouter()
	attachBucketRoutes(bucketRouter, bucketHandler, multipartHandler, objectHandler)

	// Object-related routes
	objectRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/{key:.+}`).Subrouter()
	attachObjectRoutes(objectRouter, objectHandler, multipartHandler)

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zlog.Error().Str("method", r.Method).Str("path", r.URL.Path).Msg("method not allowed")
		WriteError(w, r, MethodNotAllowedError(r))
	})

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		zlog.Info().Str("path", r.URL.Path).Msg("path not found")
		if bucketNameValidator.MatchString(r.URL.Path) {
			WriteError(w, r, NoSuchKeyError(r))
		} else {
			WriteError(w, r, InvalidBucketNameError(r))
		}
	})

	return router
}
