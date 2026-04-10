package imageutil

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/disintegration/gift"
	"github.com/gen2brain/avif"
	_ "golang.org/x/image/webp"
)

// Decode attempts to decode an image from an io.Reader, falling back to 
// safe format-aware CLI tools if the standard library fails.
func Decode(r io.Reader) (image.Image, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read data: %w", err)
	}

	// 1. Try standard decode
	img, format, err := image.Decode(bytes.NewReader(data))
	if err == nil {
		return img, format, nil
	}

	// 2. Fallback to CLI tools (specifically for AVIF/HEIC support)
	tmpInput := filepath.Join(os.TempDir(), fmt.Sprintf("imgconv_%d", time.Now().UnixNano()))
	tmpOutput := tmpInput + ".png"

	// Try with several modern extensions for tools that detect by extension
	// (though we prioritize content-based detection if tools support it)
	tmpInputExt := tmpInput + ".avif"
	
	if err := os.WriteFile(tmpInputExt, data, 0644); err != nil {
		return nil, "", fmt.Errorf("failed to write temp file: %w", err)
	}
	defer os.Remove(tmpInputExt)
	defer os.Remove(tmpOutput)

	// Fallback Path A: avifdec (specifically for AVIF)
	cmd := exec.Command("/usr/bin/avifdec", tmpInputExt, tmpOutput)
	if err := cmd.Run(); err == nil {
		if convertedImg, _, err := decodeFromFile(tmpOutput); err == nil {
			return convertedImg, "avif", nil
		}
	}

	// Fallback Path B: vips (powerful for many formats)
	cmd = exec.Command("/usr/bin/vips", "copy", tmpInputExt, tmpOutput)
	if err := cmd.Run(); err == nil {
		if convertedImg, _, err := decodeFromFile(tmpOutput); err == nil {
			return convertedImg, "vips-converted", nil
		}
	}

	return nil, "", fmt.Errorf("all decoding and conversion fallbacks failed: %w", err)
}

func decodeFromFile(path string) (image.Image, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return image.Decode(bytes.NewReader(data))
}

// Resize resizes the image to the target width and height.
// If one of them is 0, the aspect ratio is preserved based on the other side.
func Resize(img image.Image, width, height int) image.Image {
	if width == 0 && height == 0 {
		return img
	}

	g := gift.New()
	if width > 0 && height == 0 {
		g.Add(gift.Resize(width, 0, gift.LanczosResampling))
	} else if width == 0 && height > 0 {
		g.Add(gift.Resize(0, height, gift.LanczosResampling))
	} else {
		g.Add(gift.Resize(width, height, gift.LanczosResampling))
	}

	dst := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)
	return dst
}

// Fill resizes and crops the image to fill the target width and height.
func Fill(img image.Image, width, height int) image.Image {
	g := gift.New(gift.ResizeToFill(width, height, gift.LanczosResampling, gift.CenterAnchor))
	dst := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(dst, img)
	return dst
}

// EncodeAVIF encodes the image into AVIF format.
func EncodeAVIF(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	if quality <= 0 {
		quality = 70
	}
	if err := avif.Encode(&buf, img, avif.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
