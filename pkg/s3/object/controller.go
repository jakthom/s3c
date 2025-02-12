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
	// // PutObject sets an object
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (*PutObjectResult, error)
	// // DeleteObject deletes an object
	DeleteObject(r *http.Request, bucket, key, version string) (*DeleteObjectResult, error)
}

// MultipartController is an interface that specifies multipart-related
// functionality
type MultipartController interface {
	// ListMultipart lists in-progress multipart uploads in a bucket
	ListMultipart(r *http.Request, bucket, keyMarker, uploadIDMarker string, maxUploads int) (*ListMultipartResult, error)
	// InitMultipart initializes a new multipart upload
	InitMultipart(r *http.Request, bucket, key string) (string, error)
	// AbortMultipart aborts an in-progress multipart upload
	AbortMultipart(r *http.Request, bucket, key, uploadID string) error
	// CompleteMultipart finishes a multipart upload
	CompleteMultipart(r *http.Request, bucket, key, uploadID string, parts []*Part) (*CompleteMultipartResult, error)
	// ListMultipartChunks lists the constituent chunks of an in-progress
	// multipart upload
	ListMultipartChunks(r *http.Request, bucket, key, uploadID string, partNumberMarker, maxParts int) (*ListMultipartChunksResult, error)
	// UploadMultipartChunk uploads a chunk of an in-progress multipart upload
	UploadMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int, reader io.Reader) (string, error)
}
