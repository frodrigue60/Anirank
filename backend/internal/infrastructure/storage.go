package infrastructure

import (
	"context"
	"fmt"
	"log"
	"io"
	"os"
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
	GetFile(ctx context.Context, relativePath string) (io.ReadCloser, error)
	FileExists(ctx context.Context, relativePath string) (bool, error)
	ListFiles(ctx context.Context, prefix string) ([]string, error)
	GetEndpoint() string
	GetPublicURL() string
}

type S3Storage struct {
	client    *s3.Client
	bucket    string
	region    string
	endpoint  string
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
		endpoint:  endpoint,
		publicUrl: publicUrl,
	}, nil
}

// InitStorageFromEnv initializes the storage service based on environment variables.
// It supports STORAGE_TYPE=s3 (default) and STORAGE_TYPE=r2.
func InitStorageFromEnv(ctx context.Context) (StorageService, error) {
	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "s3"
	}

	var accessKey, secretKey, region, bucket, endpoint, publicUrl string

	if storageType == "r2" {
		accountID := os.Getenv("R2_ACCOUNT_ID")
		accessKey = os.Getenv("R2_ACCESS_KEY_ID")
		secretKey = os.Getenv("R2_SECRET_ACCESS_KEY")
		bucket = os.Getenv("R2_BUCKET_NAME")
		publicUrl = os.Getenv("R2_PUBLIC_URL")
		endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)
		region = "auto"
	} else {
		accessKey = os.Getenv("S3_ACCESS_KEY")
		secretKey = os.Getenv("S3_SECRET_KEY")
		region = os.Getenv("S3_REGION")
		bucket = os.Getenv("S3_BUCKET")
		endpoint = os.Getenv("S3_ENDPOINT")
		publicUrl = os.Getenv("S3_PUBLIC_URL")

		// Ensure endpoint has protocol if not set natively
		if endpoint != "" && !strings.HasPrefix(endpoint, "http") {
			endpoint = "http://" + endpoint
		}
	}

	// Sanitize public URL to remove trailing slash
	publicUrl = strings.TrimSuffix(publicUrl, "/")

	log.Printf("[STORAGE-INIT] Type=%s, Bucket=%s, Endpoint=%s, PublicURL=%s\n", storageType, bucket, endpoint, publicUrl)
	if accessKey != "" {
		maskKey := accessKey
		if len(maskKey) > 4 {
			maskKey = maskKey[:4] + "****"
		}
		log.Printf("[STORAGE-INIT] Credentials: AccessKeyID=%s\n", maskKey)
	}

	return NewS3Storage(ctx, accessKey, secretKey, region, bucket, endpoint, publicUrl)
}

func (s *S3Storage) UploadFile(ctx context.Context, relativePath string, file io.Reader, size int64, contentType string) (string, error) {
	log.Printf("[STORAGE-DEBUG] Uploading to R2/S3: Bucket=%s, Key=%s, ContentType=%s, Size=%d\n", s.bucket, relativePath, contentType, size)

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
		log.Printf("[STORAGE-ERROR] R2/S3 PutObject failed for Key=%s: %v\n", relativePath, err)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("[STORAGE-SUCCESS] Successfully uploaded Key=%s to bucket %s\n", relativePath, s.bucket)
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

func (s *S3Storage) GetFile(ctx context.Context, relativePath string) (io.ReadCloser, error) {
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(relativePath),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}
	return resp.Body, nil
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

func (s *S3Storage) GetEndpoint() string {
	return s.endpoint
}

func (s *S3Storage) GetPublicURL() string {
	return s.publicUrl
}
