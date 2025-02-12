package s2

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	s3util "github.com/jakthom/s3c/pkg/s3/util"
)

// S2 is the root struct used in the s2 library
type S2 struct {
	Auth                 AuthController
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
		Multipart:            unimplementedMultipartController{},
		maxRequestBodyLength: maxRequestBodyLength,
		readBodyTimeout:      readBodyTimeout,
	}
}

// bodyReadingMiddleware creates a middleware for reading request bodies
func (h *S2) bodyReadingMiddleware(next http.Handler) http.Handler { // TODO -> break out middleware
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentLengthStr, ok := s3util.SingleHeader(r, "Content-Length")
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
