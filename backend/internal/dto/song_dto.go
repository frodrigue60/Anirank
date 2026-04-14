package dto

type SongTypeDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

type SongMinimalDTO struct {
	ID               string        `json:"id"`
	SongRomaji       *string       `json:"song_romaji,omitempty"`
	SongEN           *string       `json:"song_en,omitempty"`
	SongJP           *string       `json:"song_jp,omitempty"`
	Slug             string        `json:"slug"`
	Type             string        `json:"type"`
	TypeID           string        `json:"type_id"`
	SongType         *SongTypeDTO  `json:"song_type,omitempty"`
	AverageRating    float64       `json:"average_rating"`
	Artists          []ArtistMinimalDTO `json:"artists,omitempty"`
	Anime            *SongAnimeDTO      `json:"anime,omitempty"`
	FavoritesCount   uint64             `json:"favorites_count"`
	Views            uint64             `json:"views"`
	UserRating       *float64           `json:"user_rating,omitempty"`
	PrevMainRank     *uint64            `json:"prev_main_rank,omitempty"`
	PrevSeasonalRank *uint64            `json:"prev_seasonal_rank,omitempty"`
	AnimeID          string             `json:"anime_id"`
	YearID           string             `json:"year_id"`
	SeasonID         string             `json:"season_id"`
}

type SongDTO struct {
	SongMinimalDTO
	LikesCount    uint64           `json:"likes_count"`
	DislikesCount uint64           `json:"dislikes_count"`
	IsFavorited   bool             `json:"is_favorited"`
	IsLiked       bool             `json:"is_liked"`
	IsDisliked    bool             `json:"is_disliked"`
	Variants      []SongVariantDTO `json:"variants,omitempty"`
}

type SongVariantDTO struct {
	ID            string  `json:"id"`
	VersionNumber uint64  `json:"version_number"`
	Slug          string  `json:"slug"`
	VideoUrl      *string `json:"video_url,omitempty"`
	Spoiler       bool    `json:"spoiler"`
}

type HomeDTO struct {
	FeaturedSong    *SongDTO           `json:"featured_song"`
	WeaklyRanking   WeaklyRankingDTO   `json:"weakly_ranking"`
	RecentlyAdded   []SongMinimalDTO   `json:"recently_added"`
	MostPopular     []SongMinimalDTO   `json:"most_popular"`
	MostViewed      []SongMinimalDTO   `json:"most_viewed"`
	FeaturedArtists []ArtistMinimalDTO `json:"featured_artists"`
}

type WeaklyRankingDTO struct {
	OP []SongMinimalDTO `json:"op"`
	ED []SongMinimalDTO `json:"ed"`
}
