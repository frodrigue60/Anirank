package admin

import (
	"strconv"
	"strings"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

var allowedVideoSources = map[string]struct{}{
	"TV": {}, "DVD": {}, "BD": {}, "WEB": {}, "LD": {},
}

var allowedVideoOverlaps = map[string]struct{}{
	"None": {}, "Overlap": {}, "Transition": {},
}

// ApplyVideoMetadataFromForm reads manual video metadata from a multipart form.
func ApplyVideoMetadataFromForm(c *fiber.Ctx, video *domain.SongVariantVideo) {
	if video == nil {
		return
	}

	if source := strings.TrimSpace(c.FormValue("source")); source != "" {
		if _, ok := allowedVideoSources[strings.ToUpper(source)]; ok {
			video.Source = strings.ToUpper(source)
		} else {
			video.Source = source
		}
	} else if video.Source == "" {
		video.Source = "TV"
	}

	if resStr := strings.TrimSpace(c.FormValue("resolution")); resStr != "" {
		if res, err := strconv.Atoi(resStr); err == nil && res >= 0 {
			video.Resolution = res
		}
	}

	video.IsNC = formBool(c, "is_nc")
	video.IsBD = formBool(c, "is_bd")
	video.IsUncensored = formBool(c, "is_uncensored")
	video.IsSubbed = formBool(c, "is_subbed")
	video.IsLyrics = formBool(c, "is_lyrics")

	if overlap := strings.TrimSpace(c.FormValue("overlap")); overlap != "" {
		if _, ok := allowedVideoOverlaps[overlap]; ok {
			video.Overlap = overlap
		} else {
			video.Overlap = overlap
		}
	} else if video.Overlap == "" {
		video.Overlap = "None"
	}

	if video.Source == "BD" {
		video.IsBD = true
	}
}

func formBool(c *fiber.Ctx, key string) bool {
	return c.FormValue(key) == "true"
}

// MetadataTargetFromForm returns how metadata should be applied.
// "new" = upcoming upload; otherwise keys like "src:<path>".
func MetadataTargetFromForm(c *fiber.Ctx) string {
	target := strings.TrimSpace(c.FormValue("metadata_target"))
	if target == "" {
		return "new"
	}
	return target
}
