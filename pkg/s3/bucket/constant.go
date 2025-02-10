package s3bucket

const (
	// defaultMaxKeys specifies the maximum number of keys returned in object
	// listings by default
	DefaultMaxKeys int = 1000
	// VersioningDisabled specifies that versioning is not enabled on a bucket
	VersioningDisabled string = ""
	// VersioningDisabled specifies that versioning is suspended on a bucket
	VersioningSuspended string = "Suspended"
	// VersioningDisabled specifies that versioning is enabled on a bucket
	VersioningEnabled string = "Enabled"
)
