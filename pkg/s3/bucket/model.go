package s3bucket

import (
	"os/user"
	"time"

	s3object "github.com/jakthom/s3c/pkg/s3/object"
)

// Bucket is an XML marshallable representation of a bucket
type Bucket struct {
	// Name is the bucket name
	Name string `xml:"Name"`
	// CreationDate is when the bucket was created
	CreationDate time.Time `xml:"CreationDate"`
}

// Version specifies a specific version of an object in a
// versioning-enabled bucket.
type Version struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string `xml:"VersionId"`
	// IsLatest specifies whether this is the latest version of the object.
	IsLatest bool `xml:"IsLatest"`
	// LastModified specifies when the object was last modified
	LastModified time.Time `xml:"LastModified"`
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string `xml:"ETag"`
	// Size specifies the size of the object
	Size uint64 `xml:"Size"`
	// StorageClass specifies the storage class used for the object
	StorageClass string `xml:"StorageClass"`
	// Owner specifies the owner of the object
	Owner user.User `xml:"Owner"`
}

// ListObjectsResult is a response from a ListObjects call
type ListObjectsResult struct {
	// Contents are the list of objects returned
	Contents []*s3object.Object
	// CommonPrefixes are the list of common prefixes returned
	CommonPrefixes []*s3object.CommonPrefixes
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
}

// ListObjectVersionsResult is a response from a ListObjectVersions call
type ListObjectVersionsResult struct {
	// Versions are the list of versions returned
	Versions []*Version
	// DeleteMarkers are the list of delete markers returned
	DeleteMarkers []*s3object.DeleteMarker
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
}
