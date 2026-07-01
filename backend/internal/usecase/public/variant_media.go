package public

import (
	"regexp"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

var iframeSrcRegex = regexp.MustCompile(`src="([^"]+)"`)

func resolveVariantVideos(v *domain.SongVariant, mediaService infrastructure.MediaService) {
	if v == nil || mediaService == nil {
		return
	}

	for j := range v.Videos {
		if v.Videos[j].EmbedUrl != nil {
			matches := iframeSrcRegex.FindStringSubmatch(*v.Videos[j].EmbedUrl)
			if len(matches) > 1 {
				v.Videos[j].EmbedUrl = &matches[1]
			}
			v.Videos[j].EmbedUrl = mediaService.Resolve(v.Videos[j].EmbedUrl)
		}
		if v.Videos[j].LocalUrl != nil {
			v.Videos[j].LocalUrl = mediaService.Resolve(v.Videos[j].LocalUrl)
		}
	}

	if len(v.Videos) > 0 {
		v.Video = &v.Videos[0]
	} else if v.Video != nil {
		if v.Video.EmbedUrl != nil {
			matches := iframeSrcRegex.FindStringSubmatch(*v.Video.EmbedUrl)
			if len(matches) > 1 {
				v.Video.EmbedUrl = &matches[1]
			}
			v.Video.EmbedUrl = mediaService.Resolve(v.Video.EmbedUrl)
		}
		if v.Video.LocalUrl != nil {
			v.Video.LocalUrl = mediaService.Resolve(v.Video.LocalUrl)
		}
	}
}

func activeVariantsForSong(variants []domain.SongVariant, mediaService infrastructure.MediaService) []domain.SongVariant {
	active := make([]domain.SongVariant, 0, len(variants))
	for _, v := range variants {
		if !v.Status {
			continue
		}
		resolved := v
		resolveVariantVideos(&resolved, mediaService)
		active = append(active, resolved)
	}
	return active
}
