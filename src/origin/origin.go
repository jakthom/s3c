package origin

import "github.com/jakthom/s3c/pkg/config"

// The future home of the origin interface

type Origin interface {
	Type() string
	Put() error
	Get() error
	List() error
}

func NewOriginFromConfig(config.Origin) Origin {
	return nil
}
