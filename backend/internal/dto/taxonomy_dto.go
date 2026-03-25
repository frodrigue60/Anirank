package dto

type StudioDTO struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	LogoUrl    *string `json:"logo_url,omitempty"`
	BannerUrl  *string `json:"banner_url,omitempty"`
	AnimeCount int     `json:"anime_count"`
}

type ProducerDTO struct {
	ID         uint64  `json:"id"`
	Name       string  `json:"name"`
	Slug       string  `json:"slug"`
	LogoUrl    *string `json:"logo_url,omitempty"`
	BannerUrl  *string `json:"banner_url,omitempty"`
	AnimeCount int     `json:"anime_count"`
}

type GenreDTO struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}
