package s3multipart

import "time"

const (
	// defaultMaxUploads specifies the maximum number of uploads returned in
	// multipart upload listings by default
	defaultMaxUploads = 1000
	// defaultMaxParts specifies the maximum number of parts returned in
	// multipart upload part listings by default
	defaultMaxParts = 1000
	// maxPartsAllowed specifies the maximum number of parts that can be
	// uploaded in a multipart upload
	maxPartsAllowed = 10000
	// completeMultipartPing is how long to wait before sending whitespace in
	// a complete multipart response (to ensure the connection doesn't close.)
	completeMultipartPing = 10 * time.Second
)
