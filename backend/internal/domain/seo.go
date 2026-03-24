package domain

// SEOData represents the metadata for a page, used for Open Graph and Twitter tags.
type SEOData struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Image       string `json:"image"`
	URL         string `json:"url"`
	Type        string `json:"type"`
	SiteName    string `json:"site_name"`
}

// SEOUsecase defines the interface for generating SEO metadata based on a path.
type SEOUsecase interface {
	GetMetadata(path string) (*SEOData, error)
}
