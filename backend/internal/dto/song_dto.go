package dto

type SongTypeDTO struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
	Description *string `json:"description,omitempty"`
}

type SongMinimalDTO struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	SongRomaji       *string       `json:"song_romaji,omitempty"`
	SongEN           *string       `json:"song_en,omitempty"`
	SongJP           *string       `json:"song_jp,omitempty"`
	Slug             string        `json:"slug"`
	Type             string        `json:"type"`
	ThemeNum         string        `json:"theme_num"`
	TypeID           string        `json:"type_id"`
	SongType         *SongTypeDTO  `json:"song_type,omitempty"`
	AverageRating    float64       `json:"average_rating"`
	Artists          []ArtistMinimalDTO `json:"artists,omitempty"`
	Anime            *SongAnimeDTO      `json:"anime,omitempty"`
	FavoritesCount   uint64             `json:"favorites_count"`
	Views            uint64             `json:"views"`
	UserRating       *float64           `json:"user_rating,omitempty"`
	PrevMainRank     *int              `json:"prev_main_rank,omitempty"`
	PrevSeasonalRank *int              `json:"prev_seasonal_rank,omitempty"`
	PrevRank         *int              `json:"prev_rank,omitempty"`
	AnimeID          string             `json:"anime_id"`
	YearID           string             `json:"year_id"`
	SeasonID         string             `json:"season_id"`
	Season           *SeasonDTO         `json:"season,omitempty"`
	Year             *YearDTO           `json:"year,omitempty"`
	Variants         []SongVariantDTO   `json:"variants,omitempty"`
}

type SongSlimDTO struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Slug          string             `json:"slug"`
	Type          string             `json:"type"`
	AverageRating float64            `json:"average_rating"`
	Artists       []ArtistSlimDTO    `json:"artists"`
	Anime         AnimeSlimDTO       `json:"anime"`
	Views         uint64             `json:"views"`
	UserRating    *float64           `json:"user_rating,omitempty"`
	Season        *SeasonDTO         `json:"season,omitempty"`
	Year          *YearDTO           `json:"year,omitempty"`
	PrevMainRank     *int            `json:"prev_main_rank,omitempty"`
	PrevSeasonalRank *int            `json:"prev_seasonal_rank,omitempty"`
	PrevRank         *int            `json:"prev_rank,omitempty"`
}

type ArtistSlimDTO struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type AnimeSlimDTO struct {
	Title    string  `json:"title"`
	Slug     string  `json:"slug"`
	CoverUrl string  `json:"cover_url"`
	BannerUrl *string `json:"banner_url"`
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

type SongVariantVideoDTO struct {
	VideoUrl     *string `json:"video_url,omitempty"`
	EmbedUrl     *string `json:"embed_url,omitempty"`
	LocalUrl     *string `json:"local_url,omitempty"`
	EmbedCode    *string `json:"embed_code,omitempty"`
	VideoSrc     *string `json:"video_src,omitempty"`
	IsNC         bool    `json:"is_nc"`
	IsBD         bool    `json:"is_bd"`
	Resolution   int     `json:"resolution"`
	IsUncensored bool    `json:"is_uncensored"`
	IsSubbed     bool    `json:"is_subbed"`
	IsLyrics     bool    `json:"is_lyrics"`
	Source       string  `json:"source"`
	Overlap      string  `json:"overlap"`
}

type SongVariantDTO struct {
	ID            string                `json:"id"`
	VersionNumber uint64                `json:"version_number"`
	Slug          string                `json:"slug"`
	VideoUrl      *string               `json:"video_url,omitempty"`
	EmbedUrl      *string               `json:"embed_url,omitempty"`
	LocalUrl      *string               `json:"local_url,omitempty"`
	EmbedCode     *string               `json:"embed_code,omitempty"`
	VideoSrc      *string               `json:"video_src,omitempty"`
	IsNC          bool                  `json:"is_nc"`
	IsBD          bool                  `json:"is_bd"`
	Resolution    int                   `json:"resolution"`
	IsUncensored  bool                  `json:"is_uncensored"`
	IsSubbed      bool                  `json:"is_subbed"`
	IsLyrics      bool                  `json:"is_lyrics"`
	Source        string                `json:"source"`
	Overlap       string                `json:"overlap"`
	Videos        []SongVariantVideoDTO `json:"videos,omitempty"`
	Episodes      *string               `json:"episodes,omitempty"`
	Spoiler       bool                  `json:"spoiler"`
	NSFW          bool                  `json:"nsfw"`
	Season        *SeasonDTO            `json:"season,omitempty"`
	Year          *YearDTO              `json:"year,omitempty"`
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
