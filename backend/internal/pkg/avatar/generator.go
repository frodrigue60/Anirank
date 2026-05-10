package avatar

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"strings"
	"unicode"
	_ "embed"

	"github.com/fogleman/gg"
	"github.com/gen2brain/avif"
	"github.com/golang/freetype/truetype"
)

//go:embed assets/font.ttf
var defaultFont []byte

// Result contains the avatar image data and metadata
type Result struct {
	Data        []byte
	Size        int64
	ContentType string
}

var palette = []string{
	"#ff4e50", "#fc913a", "#f9d423", "#ede574", "#e1f5c4", // Warm
	"#7f13ec", "#10b981", "#3b82f6", "#f59e0b", "#6366f1", // Vibrant
	"#ec4899", "#8b5cf6", "#06b6d4", "#14b8a6", "#ef4444", // Modern
}

// Generate creates an initials-based avatar locally using the gg library
func Generate(ctx context.Context, name string, size int) (*Result, error) {
	if size <= 0 {
		size = 512
	}
	dc := gg.NewContext(size, size)

	// 1. Get Initials
	initials := getInitials(name)

	// 2. Select Background Color (Deterministic based on name)
	bgColor := getBackgroundColor(name)
	dc.SetHexColor(bgColor)
	dc.Clear()

	// 3. Select Font Color (High contrast)
	if isDark(bgColor) {
		dc.SetRGB(1, 1, 1)
	} else {
		dc.SetRGB(0, 0, 0)
	}

	// 4. Load Font (Embedded)
	f, err := truetype.Parse(defaultFont)
	if err != nil {
		fmt.Printf("[Avatar] Warning: failed to parse embedded font: %v\n", err)
	} else {
		// Reduced size slightly for better visual balance (from /2 to /2.2)
		face := truetype.NewFace(f, &truetype.Options{
			Size: float64(size) / 2.2,
		})
		dc.SetFontFace(face)
	}

	// 5. Draw Text
	// We apply a visual vertical offset (-0.07 * size) to compensate for font descender metrics
	// ensuring capital initials look optically centered.
	dc.DrawStringAnchored(initials, float64(size)/2, (float64(size)/2)-(float64(size)*0.07), 0.5, 0.5)

	// 6. Encode to AVIF
	var buf bytes.Buffer
	if err := avif.Encode(&buf, dc.Image(), avif.Options{Quality: 60}); err != nil {
		return nil, fmt.Errorf("failed to encode avatar to avif: %w", err)
	}

	return &Result{
		Data:        buf.Bytes(),
		Size:        int64(buf.Len()),
		ContentType: "image/avif",
	}, nil
}

func getInitials(name string) string {
	// 1. Separate by common delimiters (underscores, dashes) and normalize
	cn := name
	cn = strings.ReplaceAll(cn, "_", " ")
	cn = strings.ReplaceAll(cn, "-", " ")
	parts := strings.Fields(cn)
	if len(parts) == 0 {
		return "?"
	}

	// 2. Extract first valid alphanumeric/letter character from each part
	var candidates []rune
	for _, p := range parts {
		for _, r := range p {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				candidates = append(candidates, unicode.ToUpper(r))
				break
			}
		}
	}

	// 3. Handle cases
	if len(candidates) == 0 {
		// Fallback for names that are ONLY symbols (e.g., "★★★")
		runes := []rune(name)
		if len(runes) > 0 {
			return strings.ToUpper(string(runes[:1]))
		}
		return "?"
	}

	// Single word case: (e.g., "Aimer", "LiSA", "澤野弘之")
	if len(candidates) == 1 {
		// Try to take up to 2 valid characters from the first word
		firstPart := parts[0]
		var wordCharacters []rune
		for _, r := range firstPart {
			if unicode.IsLetter(r) || unicode.IsDigit(r) {
				wordCharacters = append(wordCharacters, unicode.ToUpper(r))
			}
		}
		if len(wordCharacters) >= 2 {
			return string(wordCharacters[:2])
		}
		return string(wordCharacters[:1])
	}

	// Multiple words case: (e.g., "Sawano Hiroyuki")
	// Use first and last initials
	first := candidates[0]
	last := candidates[len(candidates)-1]
	return string([]rune{first, last})
}

func getBackgroundColor(name string) string {
	hash := crc32.ChecksumIEEE([]byte(name))
	index := int(hash) % len(palette)
	return palette[index]
}

func isDark(hex string) bool {
	if len(hex) < 7 {
		return false
	}
	var r, g, b uint8
	fmt.Sscanf(hex, "#%02x%02x%02x", &r, &g, &b)
	// HSP color model brightness formula
	brightness := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	return brightness < 150
}
