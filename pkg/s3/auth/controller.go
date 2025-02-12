package s3auth

// AuthController is an interface defining authentication
type AuthController interface {
	// SecretKey is called when a request is made using AWS' auth V4. If
	// the given access key exists, a non-nil secret key should be returned.
	// Otherwise nil should be returned.
	SecretKey(accessKey string, region string) (string, error)
}

type BasicAuthController struct {
	Region          string
	AccessKeyId     string
	SecretAccessKey string
}

func (b *BasicAuthController) SecretKey(accessKeyId string, region string) (string, error) {
	if accessKeyId == b.AccessKeyId && region == b.Region {
		return b.SecretAccessKey, nil
	}
	return "", nil
}
