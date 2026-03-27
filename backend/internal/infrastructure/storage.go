package infrastructure

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type StorageService interface {
	UploadFile(ctx context.Context, relativePath string, file io.Reader, size int64, contentType string) (string, error)
	GetURL(relativePath string) string
	DeleteFile(ctx context.Context, relativePath string) error
	FileExists(ctx context.Context, relativePath string) (bool, error)
	ListFiles(ctx context.Context, prefix string) ([]string, error)
}

type S3Storage struct {
	client    *s3.Client
	bucket    string
	region    string
	publicUrl string
}

func (s *S3Storage) FileExists(ctx context.Context, relativePath string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(relativePath),
	})
	if err != nil {
		// If object doesn't exist, HeadObject returns 404
		return false, nil
	}
	return true, nil
}

func NewS3Storage(ctx context.Context, accessKey, secretKey, region, bucket, endpoint, publicUrl string) (*S3Storage, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %w", err)
	}

	clientOpts := func(o *s3.Options) {
		o.UsePathStyle = true
		if endpoint != "" {
			o.BaseEndpoint = aws.String(endpoint)
		}
	}

	client := s3.NewFromConfig(cfg, clientOpts)
	return &S3Storage{
		client:    client,
		bucket:    bucket,
		region:    region,
		publicUrl: publicUrl,
	}, nil
}

func (s *S3Storage) UploadFile(ctx context.Context, relativePath string, file io.Reader, size int64, contentType string) (string, error) {
	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(relativePath),
		Body:        file,
		ContentType: aws.String(contentType),
	}

	if size > 0 {
		putInput.ContentLength = aws.Int64(size)
	}

	_, err := s.client.PutObject(ctx, putInput)
	if err != nil {
		log.Printf("ERROR: S3 PutObject failed: %v", err)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	return relativePath, nil
}

func (s *S3Storage) GetURL(relativePath string) string {
	if relativePath == "" {
		return ""
	}
	// If the path is already a full URL, return it as-is
	if strings.HasPrefix(relativePath, "http") {
		return relativePath
	}
	if s.publicUrl != "" {
		return fmt.Sprintf("%s/%s", s.publicUrl, relativePath)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, relativePath)
}

func (s *S3Storage) DeleteFile(ctx context.Context, relativePath string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(relativePath),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *S3Storage) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	var files []string
	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list files: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				files = append(files, *obj.Key)
			}
		}
	}

	return files, nil
}
