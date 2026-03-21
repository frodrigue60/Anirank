package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/h2non/bimg"
)

type MediaService interface {
	GetURL(path string) string
	Resolve(path *string) *string
	GeneratePath(prefix string, id uint64, ext string) string
	UploadImage(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string) (string, string, error)
}

type mediaService struct {
	storage StorageService
	baseURL string
}

func NewMediaService(storage StorageService) MediaService {
	// Fallback to Env if storage doesn't provide it (standardizing)
	baseURL := os.Getenv("S3_PUBLIC_URL")
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	
	return &mediaService{
		storage: storage,
		baseURL: baseURL,
	}
}

func (s *mediaService) GetURL(path string) string {
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "http") {
		return path
	}
	return s.storage.GetURL(path)
}

func (s *mediaService) Resolve(path *string) *string {
	if path == nil || *path == "" {
		return nil
	}
	url := s.GetURL(*path)
	return &url
}

func (s *mediaService) GeneratePath(prefix string, id uint64, ext string) string {
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return fmt.Sprintf("%s/%d_%s%s", prefix, id, uuid.New().String(), ext)
}

func (s *mediaService) UploadImage(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string) (string, string, error) {
	// Read the file into a buffer
	buffer, err := io.ReadAll(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to read file: %w", err)
	}

	// Process image with bimg
	// Convert to WebP, quality 80 for optimal size/quality ratio
	options := bimg.Options{
		Type:    bimg.WEBP,
		Quality: 80,
	}

	newImage, err := bimg.NewImage(buffer).Process(options)
	if err != nil {
		// If processing fails, fallback or return error
		return "", "", fmt.Errorf("failed to process image with bimg: %w", err)
	}

	// Overwrite extension and content type
	ext := "webp"
	finalContentType := "image/webp"

	filename := s.GeneratePath(prefix, id, ext)
	_, err = s.storage.UploadFile(ctx, filename, bytes.NewReader(newImage), int64(len(newImage)), finalContentType)
	if err != nil {
		return "", "", err
	}

	return filename, s.storage.GetURL(filename), nil
}
