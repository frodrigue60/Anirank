package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"anirank/api/internal/pkg/imageutil"
	"github.com/google/uuid"
)

type MediaService interface {
	GetURL(path string) string
	Resolve(path *string) *string
	GeneratePath(prefix string, id uint64, ext string) string
	UploadImage(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string) (string, string, error)
	UploadImageOptimized(ctx context.Context, prefix string, id uint64, file io.Reader, options ImageOptions) (string, string, error)
}

type ImageOptions struct {
	Width   int
	Height  int
	Format  string // "avif", "jpg", "png"
	Quality int
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
	return s.UploadImageOptimized(ctx, prefix, id, file, ImageOptions{Format: "jpg", Quality: 80})
}

func (s *mediaService) UploadImageOptimized(ctx context.Context, prefix string, id uint64, file io.Reader, opts ImageOptions) (string, string, error) {
	// 1. Decode
	img, _, err := imageutil.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode image: %w", err)
	}

	// 2. Resize if requested
	if opts.Width > 0 || opts.Height > 0 {
		img = imageutil.Resize(img, opts.Width, opts.Height)
	}

	// 3. Encode
	var processedData []byte
	ext := "jpg"
	finalContentType := "image/jpeg"

	if opts.Format == "avif" {
		ext = "avif"
		finalContentType = "image/avif"
		processedData, err = imageutil.EncodeAVIF(img, opts.Quality)
	} else if opts.Format == "png" {
		ext = "png"
		finalContentType = "image/png"
		var buf bytes.Buffer
		err = png.Encode(&buf, img)
		processedData = buf.Bytes()
	} else {
		// Default to JPEG
		var buf bytes.Buffer
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: opts.Quality})
		processedData = buf.Bytes()
	}

	if err != nil {
		return "", "", fmt.Errorf("failed to encode image: %w", err)
	}

	// 4. Upload
	filename := s.GeneratePath(prefix, id, ext)
	_, err = s.storage.UploadFile(ctx, filename, bytes.NewReader(processedData), int64(len(processedData)), finalContentType)
	if err != nil {
		return "", "", err
	}

	return filename, s.storage.GetURL(filename), nil
}
