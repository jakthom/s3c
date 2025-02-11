package fileorigin

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	s3bucket "github.com/jakthom/s3c/pkg/s3/bucket"
	s3object "github.com/jakthom/s3c/pkg/s3/object"
	s3service "github.com/jakthom/s3c/pkg/s3/service"
	"github.com/rs/zerolog/log"
)

type FileOrigin struct {
	ServiceController *FileOriginServiceController
	BucketController  *FileOriginBucketController
	ObjectController  *FileOriginObjectController
}

func NewOrigin(dataDirectory string) *FileOrigin {
	return &FileOrigin{
		ServiceController: &FileOriginServiceController{
			dataDir: dataDirectory,
		},
		BucketController: &FileOriginBucketController{
			dataDir: dataDirectory,
		},
		ObjectController: &FileOriginObjectController{
			dataDir: dataDirectory,
		},
	}
}

type FileOriginServiceController struct {
	dataDir string
}

func (c *FileOriginServiceController) ListBuckets(*http.Request) (*s3service.ListBucketsResult, error) {
	files, err := os.ReadDir(c.dataDir)
	log.Debug().Msg("Listing buckets in directory: " + c.dataDir)
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

type FileOriginBucketController struct {
	dataDir string
}

func (c *FileOriginBucketController) GetLocation(r *http.Request, bucket string) (string, error) {
	return c.dataDir, nil
}

func (c *FileOriginBucketController) ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (*s3bucket.ListObjectsResult, error) {
	directory := filepath.Join(c.dataDir, bucket, prefix)
	var objects []*s3object.Object
	var commonPrefixes []*s3object.CommonPrefixes
	result, err := os.ReadDir(directory)
	if err != nil {
		log.Error().Err(err).Msg("Failed to list objects")
		return nil, err
	}
	for _, file := range result {
		info, _ := file.Info()
		if info.IsDir() {
			prefix := s3object.CommonPrefixes{
				Prefix: file.Name() + "/", // Add trailing slash to indicate it's a directory
			}
			commonPrefixes = append(commonPrefixes, &prefix)
		} else {
			object := s3object.Object{
				Key:          file.Name(),
				LastModified: info.ModTime(),
				Size:         uint64(info.Size()),
			}
			objects = append(objects, &object)
		}
	}
	listObjectsResult := s3bucket.ListObjectsResult{}
	if objects != nil {
		listObjectsResult.Contents = objects
	}
	if commonPrefixes != nil {
		listObjectsResult.CommonPrefixes = commonPrefixes
	}
	return &listObjectsResult, nil
}

func (c *FileOriginObjectController) CopyObject(r *http.Request, srcBucket, srcKey string, getResult *s3object.GetObjectResult, destBucket, destKey string) (string, error) {
	sourcePath := filepath.Join(c.dataDir, srcBucket, srcKey)
	destinationPath := filepath.Join(c.dataDir, destBucket, destKey)
	err := copyFile(sourcePath, destinationPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy object from path: " + sourcePath + " to path: " + destinationPath)
		return "", err
	}
	return destinationPath, nil
}

func (c *FileOriginBucketController) CreateBucket(r *http.Request, bucket string) error {
	bucketDir := filepath.Join(c.dataDir, bucket)
	err := os.Mkdir(bucketDir, 0755)
	return err
}

func (c *FileOriginBucketController) DeleteBucket(r *http.Request, bucket string) error {
	bucketDir := filepath.Join(c.dataDir, bucket)
	err := os.RemoveAll(bucketDir)
	return err
}

type FileOriginObjectController struct {
	dataDir string
}

func (c *FileOriginObjectController) GetObject(r *http.Request, bucket, key, version string) (*s3object.GetObjectResult, error) {
	filePath := filepath.Join(c.dataDir, bucket, key)
	log.Info().Msg("Getting object from path: " + filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get object from path: " + filePath)
		return nil, err
	}
	info, _ := file.Stat()
	getObjectResult := s3object.GetObjectResult{
		ModTime: info.ModTime(),
		Content: file,
	}
	return &getObjectResult, nil
}

func (c *FileOriginObjectController) PutObject(r *http.Request, bucket, key string, reader io.Reader) (*s3object.PutObjectResult, error) {
	filePath := filepath.Join(c.dataDir, bucket, key)
	log.Info().Msg("Putting object to path: " + filePath)
	file, err := os.Create(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to put object to path: " + filePath)
		return nil, err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	if err != nil {
		log.Error().Err(err).Msg("Failed to copy object to path: " + filePath)
		return nil, err
	}
	return &s3object.PutObjectResult{}, nil
}

func (c *FileOriginObjectController) DeleteObject(r *http.Request, bucket, key, version string) (*s3object.DeleteObjectResult, error) {
	filePath := filepath.Join(c.dataDir, bucket, key)
	log.Info().Msg("Deleting object from path: " + filePath)
	err := os.Remove(filePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete object from path: " + filePath)
		return nil, err
	}
	return &s3object.DeleteObjectResult{}, nil
}
