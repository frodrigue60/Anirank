package public

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
)

func resolveVariantVideos(v *domain.SongVariant, mediaService infrastructure.MediaService) {
	if v == nil || mediaService == nil {
		return
	}

	storageVideos := make([]domain.SongVariantVideo, 0, len(v.Videos))
	for j := range v.Videos {
		if !domain.IsStorageVideoSrc(v.Videos[j].VideoSrc) {
			continue
		}
		// video_src is the canonical object-storage key. Older rows may not
		// have local_url populated, so always derive the public URL from it.
		v.Videos[j].LocalUrl = mediaService.Resolve(v.Videos[j].VideoSrc)
		storageVideos = append(storageVideos, v.Videos[j])
	}
	v.Videos = storageVideos

	if len(v.Videos) > 0 {
		v.Video = &v.Videos[0]
	} else {
		v.Video = nil
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
