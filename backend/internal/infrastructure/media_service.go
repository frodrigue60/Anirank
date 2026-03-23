package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"
	"io"
	"os"
	"strings"

	"github.com/disintegration/gift"
	"github.com/google/uuid"
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
	// 1. Decode Image (Supports JPEG, PNG, WebP decoding)
	img, _, err := image.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Process image with GIFT
	// For now, we just pass it through a neutral filter to normalize it, 
	// but we could add resizing or other optimizations here.
	gi := gift.New()
	dst := image.NewRGBA(gi.Bounds(img.Bounds()))
	gi.Draw(dst, img)

	// 3. Encode to JPEG (High compatibility substitute for WebP)
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80})
	if err != nil {
		return "", "", fmt.Errorf("failed to encode image to jpeg: %w", err)
	}

	processedData := buf.Bytes()

	// 4. Overwrite extension and content type
	ext := "jpg"
	finalContentType := "image/jpeg"

	filename := s.GeneratePath(prefix, id, ext)
	_, err = s.storage.UploadFile(ctx, filename, bytes.NewReader(processedData), int64(len(processedData)), finalContentType)
	if err != nil {
		return "", "", err
	}

	return filename, s.storage.GetURL(filename), nil
}
