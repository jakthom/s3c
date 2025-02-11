package s3handler

import (
	"net/http"
	"regexp"

	s3error "github.com/jakthom/s3c/pkg/s3/error"
	s3util "github.com/jakthom/s3c/pkg/s3/util"
	"github.com/rs/zerolog/log"
)

// bucketValidator is a regex for validating bucket names
var bucketValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_\.]{1,255}/`)

func NotFoundHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Info().Str("path", r.URL.Path).Msg("path not found")
		if !bucketValidator.MatchString(r.URL.Path) {
			s3util.WriteError(w, r, s3error.InvalidBucketNameError(r))
		} else {
			s3util.WriteError(w, r, s3error.NoSuchKeyError(r))
		}
	})
}
