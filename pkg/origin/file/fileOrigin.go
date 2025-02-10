package fileorigin

import (
	"net/http"
	"os"

	s3bucket "github.com/jakthom/s3c/pkg/s3/bucket"
	s3service "github.com/jakthom/s3c/pkg/s3/service"
	"github.com/rs/zerolog/log"
)

type FileOriginServiceController struct {
}

func (c *FileOriginServiceController) ListBuckets(*http.Request) (*s3service.ListBucketsResult, error) {
	directory := "./s3c/"
	files, err := os.ReadDir(directory)
	log.Debug().Msg("Listing buckets in directory: " + directory)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list buckets")
		return nil, err
	}
	var buckets []*s3bucket.Bucket

	for _, file := range files {
		if file.IsDir() {
			info, _ := file.Info()
			bucket := s3bucket.Bucket{
				Name:         file.Name(),
				CreationDate: info.ModTime(),
			}
			buckets = append(buckets, &bucket)
		}
	}
	listBucketsResult := s3service.ListBucketsResult{
		Buckets: buckets,
	}
	return &listBucketsResult, nil
}
