package s3bucket

import "net/http"

type BucketController interface {
	// GetLocation gets the location of the bucket
	GetLocation(r *http.Request, bucket string) (string, error)
	// ListObjects lists all objects within the bucket
	ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (*ListObjectsResult, error)
	// // ListObjectVersions lists all object versions within the bucket
	// ListObjectVersions(r *http.Request, bucket, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (*ListObjectVersionsResult, error)
	// CreateBucket creates a new bucket
	CreateBucket(r *http.Request, bucket string) error
	// DeleteBucket deletes the bucket
	DeleteBucket(r *http.Request, bucket string) error
	// // GetBucketVersioning gets the state of version of the bucket
	// GetBucketVersioning(r *http.Request, bucket string) (string, error)
	// // SetBucketVersioning sets the state of versioning on the bucket
	// SetBucketVersioning(r *http.Request, bucket, status string) error
}
