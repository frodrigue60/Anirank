package dto

type ArtistMinimalDTO struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	NameJP        *string `json:"name_jp,omitempty"`
	Slug          string  `json:"slug"`
	AvatarUrl     *string `json:"avatar_url"`
	EnabledSongs  int     `json:"enabled_songs"`
	DisabledSongs int     `json:"disabled_songs"`
}

type ArtistDTO struct {
	ArtistMinimalDTO
	BannerUrl      *string `json:"banner_url"`
	FavoritesCount uint64  `json:"favorites_count"`
	IsFavorited    bool    `json:"is_favorited"`
}
