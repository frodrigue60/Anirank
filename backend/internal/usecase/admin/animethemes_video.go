package admin

import (
	"strings"

	"anirank/api/internal/domain"
)

// ATVideoInput holds the minimum AnimeThemes video fields needed to build DB rows.
type ATVideoInput struct {
	Path string
	Tags string
}

// buildVideosFromATInputs maps AnimeThemes video paths and tag strings into SongVariantVideo records.
// Only metadata is stored (video_src + flags); remote .webm files are not downloaded.
func buildVideosFromATInputs(inputs []ATVideoInput) []domain.SongVariantVideo {
	videos := make([]domain.SongVariantVideo, 0, len(inputs))
	for _, entryVideo := range inputs {
		path := entryVideo.Path
		if path == "" {
			continue
		}
		isNC, isBD, resolution, isUncensored, isSubbed, isLyrics, source, overlap := parseVideoTags(entryVideo.Tags)
		vSrc := path
		if !strings.HasPrefix(strings.ToLower(vSrc), "videos/") {
			vSrc = "videos/" + vSrc
		}
		videos = append(videos, domain.SongVariantVideo{
			VideoSrc:     &vSrc,
			IsNC:         isNC,
			IsBD:         isBD,
			Resolution:   resolution,
			IsUncensored: isUncensored,
			IsSubbed:     isSubbed,
			IsLyrics:     isLyrics,
			Source:       source,
			Overlap:      overlap,
		})
	}
	return videos
}

func parseVideoTags(tags string) (isNC bool, isBD bool, resolution int, isUncensored bool, isSubbed bool, isLyrics bool, source string, overlap string) {
	normalized := strings.ToUpper(tags)
	isNC = strings.Contains(normalized, "NC")
	isBD = strings.Contains(normalized, "BD")
	isUncensored = strings.Contains(normalized, "UNCENSORED")
	isSubbed = strings.Contains(normalized, "SUBBED")
	isLyrics = strings.Contains(normalized, "LYRICS")

	if strings.Contains(normalized, "BD") || strings.Contains(normalized, "BLU-RAY") {
		source = "BD"
	} else if strings.Contains(normalized, "WEB") {
		source = "WEB"
	} else if strings.Contains(normalized, "DVD") {
		source = "DVD"
	} else if strings.Contains(normalized, "LD") {
		source = "LD"
	} else {
		source = "TV"
	}

	if strings.Contains(normalized, "OVERLAP") {
		overlap = "Overlap"
	} else if strings.Contains(normalized, "TRANSITION") {
		overlap = "Transition"
	} else {
		overlap = "None"
	}

	if strings.Contains(normalized, "2160") {
		resolution = 2160
	} else if strings.Contains(normalized, "1440") {
		resolution = 1440
	} else if strings.Contains(normalized, "1080") {
		resolution = 1080
	} else if strings.Contains(normalized, "720") {
		resolution = 720
	} else if strings.Contains(normalized, "576") {
		resolution = 576
	} else if strings.Contains(normalized, "480") {
		resolution = 480
	} else if strings.Contains(normalized, "360") {
		resolution = 360
	}
	return
}
