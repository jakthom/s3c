package s3auth

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/gorilla/mux"
	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

var authV4HeaderValidator = regexp.MustCompile(`^AWS4-HMAC-SHA256 Credential=([^/]*)/([^/]*)/([^/]*)/s3/aws4_request, ?SignedHeaders=([^,]+), ?Signature=(.+)$`)

func AuthV4(w http.ResponseWriter, r *http.Request, authController AuthController, authorizationHeader string) error {
	// Ensure the Authorization header is well-formed
	match := authV4HeaderValidator.FindStringSubmatch(authorizationHeader)
	if len(match) == 0 {
		// The Authorization header is malformed
		return s3error.AuthorizationHeaderMalformedError(r)
	}
	// Get components from Authorization header
	accessKey := match[1]
	date := match[2]
	region := match[3]
	signedHeaderKeys := strings.Split(match[4], ";")
	sort.Strings(signedHeaderKeys)
	expectedSignature := match[5]
	// Get the expected secret key
	secretKey, err := authController.SecretKey(accessKey, region)
	// Short-circuit if the secret key is not found
	// or if there is an error
	if secretKey == "" {
		return s3error.InvalidAccessKeyIDError(r)
	}
	if err != nil {
		return s3error.InternalError(r, err)
	}
	// Build the canonical request
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
		s3util.NormURI(r.URL.Path),
		s3util.NormQuery(r.URL.Query()),
		signedHeaders.String(),
		strings.Join(signedHeaderKeys, ";"),
		r.Header.Get("x-amz-content-sha256"),
	}, "\n")

	timestamp, err := s3util.ParseAWSTimestamp(r)
	if err != nil {
		return err
	}
	formattedTimestamp := s3util.FormatAWSTimestamp(timestamp)

	// step 2: construct the string to sign
	stringToSign := fmt.Sprintf(
		"AWS4-HMAC-SHA256\n%s\n%s/%s/s3/aws4_request\n%x",
		formattedTimestamp,
		date,
		region,
		sha256.Sum256([]byte(canonicalRequest)),
	)

	// step 3: calculate the signing key
	dateKey := s3util.HmacSHA256([]byte("AWS4"+secretKey), date)
	dateRegionKey := s3util.HmacSHA256(dateKey, region)
	dateRegionServiceKey := s3util.HmacSHA256(dateRegionKey, "s3")
	signingKey := s3util.HmacSHA256(dateRegionServiceKey, "aws4_request")

	// step 4: construct & verify the signature
	signature := s3util.HmacSHA256(signingKey, stringToSign)

	if expectedSignature != fmt.Sprintf("%x", signature) {
		return s3error.SignatureDoesNotMatchError(r)
	}

	vars := mux.Vars(r)
	vars["authMethod"] = "v4"
	vars["authAccessKey"] = accessKey
	vars["authRegion"] = region
	// store signature data as vars, since it may be reused for verifying chunked uploads
	vars["authSignature"] = expectedSignature
	vars["authSignatureKey"] = string(signingKey)
	vars["authSignatureTimestamp"] = formattedTimestamp
	vars["authSignatureDate"] = date
	vars["authSignatureRegion"] = region
	return nil
}
