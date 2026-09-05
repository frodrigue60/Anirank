package rate

import (
	"errors"
	"strings"

	"anirank/api/internal/domain"
)

var (
	ErrRateSongInactive  = errors.New("rate: song inactive")
	ErrRateAnimeInactive = errors.New("rate: anime inactive")
	ErrRateSongNoMedia   = errors.New("rate: song has no playable video")
)

type mediaURLResolver interface {
	GetURL(path string) string
}

// validateSongForRate enforces public catalog status and playable media.
func validateSongForRate(song *domain.Song, audioURL string) error {
	if song == nil {
		return domain.ErrNotFound
	}
	if !song.Status {
		return ErrRateSongInactive
	}
	if song.Anime == nil || !song.Anime.Status {
		return ErrRateAnimeInactive
	}
	if strings.TrimSpace(audioURL) == "" {
		return ErrRateSongNoMedia
	}
	return nil
}

func rateSongLoadErrorMessage(err error) string {
	if err == nil {
		return "Song not found"
	}
	switch {
	case errors.Is(err, ErrRateSongInactive), errors.Is(err, ErrRateAnimeInactive):
		return "This theme is inactive or not publicly available"
	case errors.Is(err, ErrRateSongNoMedia):
		return "This theme has no playable video"
	default:
		return "Song not found"
	}
}

func mediaURLFromVariant(ms mediaURLResolver, v *domain.SongVariant) string {
	if ms == nil || v == nil {
		return ""
	}

	vids := make([]*domain.SongVariantVideo, 0, 1+len(v.Videos))
	if v.Video != nil {
		vids = append(vids, v.Video)
	}
	for i := range v.Videos {
		vids = append(vids, &v.Videos[i])
	}

	for _, vid := range vids {
		if vid == nil {
			continue
		}
		if vid.LocalUrl != nil {
			if u := strings.TrimSpace(ms.GetURL(*vid.LocalUrl)); u != "" {
				return u
			}
		}
		if vid.VideoSrc != nil {
			if u := strings.TrimSpace(ms.GetURL(*vid.VideoSrc)); u != "" {
				return u
			}
		}
	}
	return ""
}

// resolvePlayableAudioURL returns the first resolvable media URL, preferring active variants.
func resolvePlayableAudioURL(ms mediaURLResolver, song *domain.Song) string {
	if song == nil || len(song.Variants) == 0 {
		return ""
	}

	for i := range song.Variants {
		v := &song.Variants[i]
		if !v.Status {
			continue
		}
		if u := mediaURLFromVariant(ms, v); u != "" {
			return u
		}
	}

	// Fallback: inactive variants that still have media (legacy rows).
	for i := range song.Variants {
		if u := mediaURLFromVariant(ms, &song.Variants[i]); u != "" {
			return u
		}
	}
	return ""
}
