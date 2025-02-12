package s3object

import (
	"io"
	"os/user"
	"time"

	s3user "github.com/jakthom/s3c/pkg/s3/user"
)

// Object is an individual file/object
type Object struct {
	// Key specifies the object key
	Key string `xml:"Key"`
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

// DeleteMarker specifies an object that has been deleted from a
// versioning-enabled bucket.
type DeleteMarker struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string `xml:"VersionId"`
	// IsLatest specifies whether this is the latest version of the object.
	IsLatest bool `xml:"IsLatest"`
	// LastModified specifies when the object was last modified
	LastModified time.Time `xml:"LastModified"`
	// Owner specifies the owner of the object
	Owner user.User `xml:"Owner"`
}

// CommonPrefixes specifies a common prefix of S3 keys. This is akin to a
// directory.
type CommonPrefixes struct {
	// Prefix specifies the common prefix value.
	Prefix string `xml:"Prefix"`
	// Owner specifies the owner of the object
	Owner user.User `xml:"Owner"`
}

// GetObjectResult is a response from a GetObject call
type GetObjectResult struct {
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
	// DeleteMarker specifies whether there's a delete marker in place of the
	// object.
	DeleteMarker bool
	// ModTime specifies when the object was modified.
	ModTime time.Time
	// Content is the contents of the object.
	Content io.ReadSeeker
}

// PutObjectResult is a response from a PutObject call
type PutObjectResult struct {
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
}

// DeleteObjectResult is a response from a DeleteObject call
type DeleteObjectResult struct {
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
	// DeleteMarker specifies whether there's a delete marker in place of the
	// object.
	DeleteMarker bool
}

// ############################################################################################################
// Multipart
// ############################################################################################################

// Upload is an XML marshallable representation of an in-progress multipart upload
type Upload struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// UploadID is an ID identifying the multipart upload
	UploadID string `xml:"UploadId"`
	// Initiator is the user that initiated the multipart upload
	Initiator s3user.User `xml:"Initiator"`
	// Owner specifies the owner of the object
	Owner s3user.User `xml:"Owner"`
	// StorageClass specifies the storage class used for the object
	StorageClass string `xml:"StorageClass"`
	// Initiated is a timestamp specifying when the multipart upload was
	// started
	Initiated time.Time `xml:"Initiated"`
}

// Part is an XML marshallable representation of a chunk of an in-progress
// multipart upload
type Part struct {
	// PartNumber is the index of the part
	PartNumber int `xml:"PartNumber"`
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string `xml:"ETag"`
}

// ListMultipartResult is a response from a ListMultipart call
type ListMultipartResult struct {
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
	// Uploads are the list of uploads returned
	Uploads []*Upload
}

// CompleteMultipartResult is a response from a CompleteMultipart call
type CompleteMultipartResult struct {
	// Location is the location of the newly uploaded object
	Location string
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
}

// ListMultipartChunksResult is a response from a ListMultipartChunks call
type ListMultipartChunksResult struct {
	// Initiator is the user that initiated the multipart upload
	Initiator *s3user.User
	// Owner specifies the owner of the object
	Owner *s3user.User
	// StorageClass specifies the storage class used for the object
	StorageClass string
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
	// Parts are the list of parts returned
	Parts []*Part
}
