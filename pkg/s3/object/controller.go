package s3object

import (
	"io"
	"net/http"
)

// ObjectController is an interface that specifies object-level functionality.
type ObjectController interface {
	// GetObject gets an object
	GetObject(r *http.Request, bucket, key, version string) (*GetObjectResult, error)
	// CopyObject copies an object
	CopyObject(r *http.Request, srcBucket, srcKey string, getResult *GetObjectResult, destBucket, destKey string) (string, error)
	// PutObject sets an object
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (*PutObjectResult, error)
	// DeleteObject deletes an object
	DeleteObject(r *http.Request, bucket, key, version string) (*DeleteObjectResult, error)
}
