package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"

	"github.com/rs/zerolog/log"
)

type responseCapture struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (rc *responseCapture) Write(b []byte) (int, error) {
	rc.body.Write(b)
	return rc.ResponseWriter.Write(b)
}

func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
	rc.ResponseWriter.WriteHeader(statusCode)
}

func DebugMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture the request
		proxyReqDump, _ := httputil.DumpRequest(r, false)
		// Write request to stdout
		fmt.Println(string(proxyReqDump))
		// Capture the response
		rc := &responseCapture{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rc, r)
		// Create a dummy response to use with DumpResponse
		dummyResp := &http.Response{
			StatusCode: rc.statusCode,
			Header:     rc.Header(),
			Body:       io.NopCloser(bytes.NewBuffer(rc.body.Bytes())),
		}
		// Dump the response
		dumpResp, err := httputil.DumpResponse(dummyResp, true)
		if err != nil {
			log.Error().Err(err).Msg("could not dump response")
			return
		}
		// Write response to stdout
		fmt.Println(string(dumpResp))
	})
}
