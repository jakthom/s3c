package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type request struct {
	// ResponseCode             int           `json:"responseCode"`
	// RequestDuration          time.Duration `json:"requestDuration"`
	// RequestDurationForHumans string        `json:"requestDurationForHumans"`
	Headers           http.Header       `json:"headers"`
	ClientIp          string            `json:"clientIp"`
	RequestMethod     string            `json:"requestMethod"`
	RequestUri        string            `json:"requestUri"`
	DecodedRequestUri string            `json:"decodedRequestUri"`
	Body              interface{}       `json:"body"`
	Vars              map[string]string `json:"vars"`
	// TODO -> Add headers
}

// getIp returns the client's IP address from the request. It first checks for
// the X-Forwarded-For header, then the X-Real-IP header, and finally falls back
// to the RemoteAddr field of the request.
func getIp(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if len(ip) == 0 {
		ip = r.Header.Get("X-Real-IP")
	}
	if len(ip) == 0 {
		ip = r.RemoteAddr
	}
	if strings.Contains(ip, ",") {
		ip = strings.Split(ip, ",")[0]
	}
	return ip
}

// GetDuration returns the associated duration between two times
func getDuration(start time.Time, end time.Time) time.Duration {
	return end.Sub(start)
}

func RequestLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf, _ := io.ReadAll(r.Body)
		r1 := io.NopCloser(bytes.NewBuffer(buf))
		r2 := io.NopCloser(bytes.NewBuffer(buf))
		reqBody, err := io.ReadAll(r1)
		r.Body = r2
		if err != nil {
			log.Error().Err(err).Msg("could not read request body")
		}

		var b interface{}
		if string(reqBody) != "" {
			err = json.Unmarshal(reqBody, &b)

			if err != nil {
				log.Debug().Err(err).Interface("body", reqBody).Msg("could not unmarshal request body")
			}
		}

		decodedUri, _ := url.QueryUnescape(r.RequestURI)
		vars := mux.Vars(r)
		req := request{
			Headers:           r.Header,
			ClientIp:          getIp(r),
			RequestMethod:     r.Method,
			RequestUri:        r.RequestURI,
			DecodedRequestUri: decodedUri,
			Body:              b,
			Vars:              vars,
		}
		log.Debug().Interface("request", req).Msg("request")
		// TODO -> Output this to a file or something, to introspect workloads and request volumes
		next.ServeHTTP(w, r)
	})
}
