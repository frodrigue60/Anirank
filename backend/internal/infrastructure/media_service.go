package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/pkg/imageutil"
	"github.com/google/uuid"
)

type MediaService interface {
	GetURL(path string) string
	Resolve(path *string) *string
	GeneratePath(prefix string, id uint64, ext string) string
	UploadImage(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string) (string, string, error)
	UploadImageOptimized(ctx context.Context, prefix string, id uint64, file io.Reader, options ImageOptions) (string, string, error)
	UploadWithResolutions(ctx context.Context, prefix string, id uint64, file io.Reader, preset ResolutionPreset) (string, string, error)
	GetImageSources(path string) []domain.ImageSource
}

type ResolutionPreset string

const (
	PresetSquare    ResolutionPreset = "square"    // 128, 400
	PresetPoster    ResolutionPreset = "poster"    // 240x360, 600x900
	PresetLandscape ResolutionPreset = "landscape" // 640x360, 1280x720
	PresetAnnouncement ResolutionPreset = "announcement" // 450
)

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

func (s *mediaService) UploadWithResolutions(ctx context.Context, prefix string, id uint64, file io.Reader, preset ResolutionPreset) (string, string, error) {
	// 1. Decode original
	img, _, err := imageutil.Decode(file)
	if err != nil {
		return "", "", fmt.Errorf("failed to decode original: %w", err)
	}

	// Optimize storage: For announcements, we don't want high-res originals.
	// Resize immediately if it exceeds target width.
	if preset == PresetAnnouncement && img.Bounds().Dx() > 450 {
		img = imageutil.Resize(img, 450, 0)
	}

	// Define resolutions based on preset
	var resolutions []int
	isSquare := preset == PresetSquare

	switch preset {
	case PresetSquare:
		resolutions = []int{128, 400}
	case PresetPoster:
		resolutions = []int{240, 600}
	case PresetLandscape:
		resolutions = []int{640, 1280}
	case PresetAnnouncement:
		resolutions = []int{450}
	default:
		resolutions = []int{600}
	}

	// 2. Upload Fallback (as JPEG for compatibility)
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return "", "", fmt.Errorf("failed to encode fallback: %w", err)
	}
	originalPath := s.GeneratePath(prefix, id, "jpg")
	_, err = s.storage.UploadFile(ctx, originalPath, bytes.NewReader(buf.Bytes()), int64(buf.Len()), "image/jpeg")
	if err != nil {
		return "", "", fmt.Errorf("failed to upload fallback image: %w", err)
	}

	// 3. Generate and Upload AVIF versions
	// Path example: path/to/image.jpg -> path/to/image_sm.avif
	pathWithoutExt := strings.TrimSuffix(originalPath, ".jpg")

	for _, w := range resolutions {
		var suffix string
		var resized image.Image

		// Logic for suffixes and resizing
		if w <= 300 {
			suffix = "_sm"
		} else if w <= 800 {
			suffix = "_md"
		} else {
			suffix = "_lg"
		}

		if isSquare {
			resized = imageutil.Fill(img, w, w)
		} else if preset == PresetPoster {
			resized = imageutil.Resize(img, w, int(float64(w)*1.5))
		} else {
			// Only downscale if original is larger than target width
			if img.Bounds().Dx() > w {
				resized = imageutil.Resize(img, w, 0) // Preserve aspect ratio
			} else {
				resized = img
			}
		}

		avifData, err := imageutil.EncodeAVIF(resized, 65) // Quality 65 for AVIF is very good
		if err != nil {
			continue // Skip failed resolutions but keep going
		}

		resPath := pathWithoutExt + suffix + ".avif"
		_, _ = s.storage.UploadFile(ctx, resPath, bytes.NewReader(avifData), int64(len(avifData)), "image/avif")
	}

	return originalPath, s.storage.GetURL(originalPath), nil
}

func (s *mediaService) GetImageSources(path string) []domain.ImageSource {
	if path == "" || strings.HasPrefix(path, "http") {
		return nil
	}

	// Deterministic mapping: we expect _sm, _md, _lg to exist if the original is there
	pathWithoutExt := strings.TrimSuffix(path, ".jpg")

	// Check what kind of path it is to determine width labels
	isSquare := strings.Contains(path, "avatars") || strings.Contains(path, "users")
	isPoster := strings.Contains(path, "covers")

	sources := []domain.ImageSource{}

	// Define expected sizes based on path
	var sizes map[string]int
	if isSquare {
		sizes = map[string]int{"_sm": 128, "_md": 400}
	} else if isPoster {
		sizes = map[string]int{"_sm": 240, "_md": 600}
	} else if strings.Contains(path, "announcements") {
		sizes = map[string]int{"_md": 450}
	} else {
		sizes = map[string]int{"_md": 640, "_lg": 1280}
	}

	for suffix, width := range sizes {
		sources = append(sources, domain.ImageSource{
			URL:   s.storage.GetURL(pathWithoutExt + suffix + ".avif"),
			Width: width,
		})
	}

	return sources
}
