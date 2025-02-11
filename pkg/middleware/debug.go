package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
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
		fmt.Println(string(proxyReqDump))
		// Capture the response
		rc := &responseCapture{ResponseWriter: w, body: &bytes.Buffer{}}
		next.ServeHTTP(rc, r)

		// Create a dummy response to use with DumpResponse
		// dummyResp := &http.Response{
		// 	StatusCode: rc.statusCode,
		// 	Header:     rc.Header(),
		// 	Body:       io.NopCloser(bytes.NewBuffer(rc.body.Bytes())),
		// }

		// // Dump the response
		// dumpResp, err := httputil.DumpResponse(dummyResp, true)
		// if err != nil {
		// 	fmt.Println("could not dump response:", err)
		// 	return
		// }

		// Log the response
		// fmt.Println(string(dumpResp))
	})
}
