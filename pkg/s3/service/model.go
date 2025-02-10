package s3service

import (
	"encoding/xml"

	s3bucket "github.com/jakthom/s3c/pkg/s3/bucket"
	s3user "github.com/jakthom/s3c/pkg/user"
)

// ListBucketsResult is a response from a ListBucket call
type ListBucketsResult struct {
	XMLName xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListAllMyBucketsResult"`
	// Owner is the owner of the buckets
	Owner *s3user.User `xml:"Owner"`
	// Buckets are a list of buckets under the given owner
	Buckets []*s3bucket.Bucket `xml:"Buckets>Bucket"`
}
