package dto

import "anirank/api/internal/domain"

type VideoAuditMissingItemDTO struct {
	VideoSrc    string `json:"video_src"`
	VariantUUID string `json:"variant_uuid"`
	VariantSlug string `json:"variant_slug"`
	SongUUID    string `json:"song_uuid"`
	SongTitle   string `json:"song_title"`
	AnimeSlug   string `json:"anime_slug"`
	AnimeTitle  string `json:"anime_title"`
}

type VideoAuditReportDTO struct {
	JobID          string                     `json:"job_id"`
	Status         string                     `json:"status"`
	Prefix         string                     `json:"prefix"`
	IncludeOrphans bool                       `json:"include_orphans"`
	TotalRows      int                        `json:"total_rows"`
	UniquePaths    int                        `json:"unique_paths"`
	PresentCount   int                        `json:"present_count"`
	MissingCount   int                        `json:"missing_count"`
	OrphanCount    int                        `json:"orphan_count"`
	Missing        []VideoAuditMissingItemDTO `json:"missing"`
	Orphans        []string                   `json:"orphans,omitempty"`
}

func ToVideoAuditReportDTO(r *domain.VideoAuditReport) VideoAuditReportDTO {
	if r == nil {
		return VideoAuditReportDTO{}
	}
	dto := VideoAuditReportDTO{
		JobID:          r.JobID,
		Status:         string(r.Status),
		Prefix:         r.Prefix,
		IncludeOrphans: r.IncludeOrphans,
		TotalRows:      r.TotalRows,
		UniquePaths:    r.UniquePaths,
		PresentCount:   r.PresentCount,
		MissingCount:   r.MissingCount,
		OrphanCount:    r.OrphanCount,
		Missing:        make([]VideoAuditMissingItemDTO, 0, len(r.Missing)),
		Orphans:        r.Orphans,
	}
	for _, m := range r.Missing {
		dto.Missing = append(dto.Missing, VideoAuditMissingItemDTO{
			VideoSrc:    m.VideoSrc,
			VariantUUID: m.VariantUUID,
			VariantSlug: m.VariantSlug,
			SongUUID:    m.SongUUID,
			SongTitle:   m.SongTitle,
			AnimeSlug:   m.AnimeSlug,
			AnimeTitle:  m.AnimeTitle,
		})
	}
	return dto
}
