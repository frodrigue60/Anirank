package domain

// VideoAuditCandidate is a DB row eligible for R2/S3 existence checking.
type VideoAuditCandidate struct {
	VideoSrc     string
	VariantUUID  string
	VariantSlug  string
	SongUUID     string
	SongTitle    string
	AnimeSlug    string
	AnimeTitle   string
}

// VideoAuditMissingItem describes a video reference in DB with no object in storage.
type VideoAuditMissingItem struct {
	VideoSrc     string `json:"video_src"`
	VariantUUID  string `json:"variant_uuid"`
	VariantSlug  string `json:"variant_slug"`
	SongUUID     string `json:"song_uuid"`
	SongTitle    string `json:"song_title"`
	AnimeSlug    string `json:"anime_slug"`
	AnimeTitle   string `json:"anime_title"`
}

// VideoAuditReport is the full result of a storage audit job.
type VideoAuditReport struct {
	JobID          string                  `json:"job_id"`
	Status         ImportJobStatus         `json:"status"`
	Prefix         string                  `json:"prefix"`
	IncludeOrphans bool                    `json:"include_orphans"`
	TotalRows      int                     `json:"total_rows"`
	UniquePaths    int                     `json:"unique_paths"`
	PresentCount   int                     `json:"present_count"`
	MissingCount   int                     `json:"missing_count"`
	OrphanCount    int                     `json:"orphan_count"`
	Missing        []VideoAuditMissingItem `json:"missing"`
	Orphans        []string                `json:"orphans,omitempty"`
}

// VideoAuditFilters scopes which video_src rows are audited.
type VideoAuditFilters struct {
	Prefix string
}
