package s3auth

import "regexp"

var authV4HeaderValidator = regexp.MustCompile(`^AWS4-HMAC-SHA256 Credential=([^/]*)/([^/]*)/([^/]*)/s3/aws4_request, ?SignedHeaders=([^,]+), ?Signature=(.+)$`)
