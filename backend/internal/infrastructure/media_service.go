package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
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
	UploadVideo(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string, originalName string) (string, string, error)
	UploadImageOptimized(ctx context.Context, prefix string, id uint64, file io.Reader, options ImageOptions) (string, string, error)
	UploadWithResolutions(ctx context.Context, prefix string, id uint64, file io.Reader, preset ResolutionPreset) (string, string, error)
	GetImageSources(path string) []domain.ImageSource
	GetFile(ctx context.Context, path string) (io.ReadCloser, error)
	DeleteMedia(ctx context.Context, path string)
	FileExists(ctx context.Context, path string) (bool, error)
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
	baseURL := storage.GetPublicURL()
	if baseURL != "" && !strings.HasSuffix(baseURL, "/") {
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

func (s *mediaService) UploadVideo(ctx context.Context, prefix string, id uint64, file io.Reader, size int64, contentType string, originalName string) (string, string, error) {
	// Sanitize original name if provided, otherwise generate one
	var filename string
	if originalName != "" {
		// Basic sanitization: replace spaces with hyphens, remove weird chars
		cleanName := strings.ReplaceAll(originalName, " ", "-")
		// Use the prefix (folder) and the clean original name directly
		filename = fmt.Sprintf("%s/%s", prefix, cleanName)
	} else {
		ext := "mp4"
		if contentType == "video/webm" {
			ext = "webm"
		}
		filename = s.GeneratePath(prefix, id, ext)
	}

	// Check if file already exists to avoid redundant uploads
	if exists, _ := s.FileExists(ctx, filename); exists {
		log.Printf("[MEDIA-DEBUG] Video already exists at %s, skipping upload\n", filename)
		return filename, s.storage.GetURL(filename), nil
	}

	log.Printf("[MEDIA-DEBUG] Uploading video: %s (%d bytes, %s)\n", filename, size, contentType)
	_, err := s.storage.UploadFile(ctx, filename, file, size, contentType)
	if err != nil {
		log.Printf("[MEDIA-ERROR] Video upload failed for %s: %v\n", filename, err)
		return "", "", err
	}

	return filename, s.storage.GetURL(filename), nil
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
	log.Printf("[MEDIA-DEBUG] Uploading processed image: %s (%d bytes, %s)\n", filename, len(processedData), finalContentType)
	_, err = s.storage.UploadFile(ctx, filename, bytes.NewReader(processedData), int64(len(processedData)), finalContentType)
	if err != nil {
		log.Printf("[MEDIA-ERROR] Upload failed for %s: %v\n", filename, err)
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

	// 2. Upload Base Image (as AVIF)
	avifData, err := imageutil.EncodeAVIF(img, 80)
	if err != nil {
		return "", "", fmt.Errorf("failed to encode base image: %w", err)
	}
	originalPath := s.GeneratePath(prefix, id, "avif")
	log.Printf("[MEDIA-DEBUG] Uploading base image: %s (%d bytes)\n", originalPath, len(avifData) )
	_, err = s.storage.UploadFile(ctx, originalPath, bytes.NewReader(avifData), int64(len(avifData)), "image/avif")
	if err != nil {
		log.Printf("[MEDIA-ERROR] Base image upload failed for %s: %v\n", originalPath, err)
		return "", "", fmt.Errorf("failed to upload base image: %w", err)
	}

	// 3. Generate and Upload AVIF versions
	// Path example: path/to/image.avif -> path/to/image_sm.avif
	pathWithoutExt := strings.TrimSuffix(originalPath, ".avif")

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
		} else {
			// Base measurement: Width, preserving aspect ratio
			if img.Bounds().Dx() > w {
				resized = imageutil.Resize(img, w, 0)
			} else {
				resized = img
			}
		}

		avifData, err := imageutil.EncodeAVIF(resized, 65) // Quality 65 for AVIF is very good
		if err != nil {
			continue // Skip failed resolutions but keep going
		}

		resPath := pathWithoutExt + suffix + ".avif"
		log.Printf("[MEDIA-DEBUG] Uploading variant: %s (%d bytes)\n", resPath, len(avifData))
		if _, err := s.storage.UploadFile(ctx, resPath, bytes.NewReader(avifData), int64(len(avifData)), "image/avif"); err != nil {
			log.Printf("[MEDIA-ERROR] Variant upload failed for %s: %v\n", resPath, err)
		}
	}

	return originalPath, s.storage.GetURL(originalPath), nil
}

func (s *mediaService) GetImageSources(path string) []domain.ImageSource {
	if path == "" || strings.HasPrefix(path, "http") {
		return nil
	}

	// Deterministic mapping: we expect _sm, _md, _lg to exist if the original is there
	pathWithoutExt := path
	if strings.HasSuffix(path, ".jpg") {
		pathWithoutExt = strings.TrimSuffix(path, ".jpg")
	} else if strings.HasSuffix(path, ".avif") {
		pathWithoutExt = strings.TrimSuffix(path, ".avif")
	}

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

func (s *mediaService) GetFile(ctx context.Context, path string) (io.ReadCloser, error) {
	if path == "" {
		return nil, fmt.Errorf("empty path")
	}

	// If it's a full URL, we might need to fetch it via HTTP if it's external,
	// or get it from storage if it's internal.
	if strings.HasPrefix(path, "http") {
		if s.baseURL != "" && strings.HasPrefix(path, s.baseURL) {
			path = strings.TrimPrefix(path, s.baseURL)
		} else {
			// External URL
			resp, err := http.Get(path)
			if err != nil {
				return nil, err
			}
			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				return nil, fmt.Errorf("failed to fetch external image: %s", resp.Status)
			}
			return resp.Body, nil
		}
	}

	return s.storage.GetFile(ctx, path)
}

func (s *mediaService) DeleteMedia(ctx context.Context, path string) {
	if path == "" {
		return
	}

	// Clean up full URLs to get relative path if needed, or ignore external ones
	if strings.HasPrefix(path, "http") {
		if !strings.HasPrefix(path, s.baseURL) {
			return // External URL, don't delete
		}
		path = strings.TrimPrefix(path, s.baseURL)
	}

	// We launch a goroutine to prevent the upload request from hanging
	// while we issue multiple AWS S3 delete commands.
	go func() {
		// Use a detached background context because the original request ctx will be cancelled
		bgCtx := context.Background()

		log.Printf("[MEDIA-DEBUG] Deleting orphan media: %s\n", path)
		_ = s.storage.DeleteFile(bgCtx, path)

		// Also try to delete standard resolution variants
		pathWithoutExt := path
		if strings.HasSuffix(path, ".jpg") {
			pathWithoutExt = strings.TrimSuffix(path, ".jpg")
		} else if strings.HasSuffix(path, ".avif") {
			pathWithoutExt = strings.TrimSuffix(path, ".avif")
		}

		variants := []string{"_sm", "_md", "_lg"}
		for _, v := range variants {
			_ = s.storage.DeleteFile(bgCtx, pathWithoutExt+v+".avif")
			_ = s.storage.DeleteFile(bgCtx, pathWithoutExt+v+".jpg")
		}
	}()
}

func (s *mediaService) FileExists(ctx context.Context, path string) (bool, error) {
	if path == "" {
		return false, nil
	}
	if strings.HasPrefix(path, "http") {
		if !strings.HasPrefix(path, s.baseURL) {
			return false, nil // External URL
		}
		path = strings.TrimPrefix(path, s.baseURL)
	}
	return s.storage.FileExists(ctx, path)
}
