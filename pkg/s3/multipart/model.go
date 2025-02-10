package s3multipart

import (
	"time"

	s3user "github.com/jakthom/s3c/pkg/user"
)

// Upload is an XML marshallable representation of an in-progress multipart
// upload
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
