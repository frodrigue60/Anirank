package dto

type ArtistMinimalDTO struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	NameJP        *string `json:"name_jp,omitempty"`
	Slug          string  `json:"slug"`
	AvatarUrl     *string `json:"avatar_url,omitempty"`
	SongsCount    int     `json:"songs_count"`
	EnabledSongs  int     `json:"enabled_songs"`
	DisabledSongs int     `json:"disabled_songs"`
	Status        bool    `json:"status"`
}

type ArtistDTO struct {
	ArtistMinimalDTO
	BannerUrl      *string `json:"banner_url,omitempty"`
	FavoritesCount uint64  `json:"favorites_count"`
	IsFavorited    bool    `json:"is_favorited"`
}
