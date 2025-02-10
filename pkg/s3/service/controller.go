package s3service

import "net/http"

// ServiceController is an interface defining service-level functionality
type ServiceController interface {
	// ListBuckets lists all buckets
	ListBuckets(r *http.Request) (*ListBucketsResult, error)
}
