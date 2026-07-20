package animethemes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.animethemes.moe"

// AnimeThemesClient defines the interface for the AnimeThemes API.
type AnimeThemesClient interface {
	FetchAnimePage(ctx context.Context, page, pageSize int) (*AnimePageResponse, error)
	FetchSongPage(ctx context.Context, page, pageSize int, idGt uint64) (*SongPageResponse, error)
	FetchThemeByID(ctx context.Context, themeID uint64) (*ATTheme, error)
	FetchAnimeBySlug(ctx context.Context, slug string) (*ATAnime, error)
}

// Client is the concrete HTTP client for api.animethemes.moe.
type Client struct {
	httpClient *http.Client
}

// NewClient creates a new AnimeThemes API client.
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ─── Response Types ──────────────────────────────────────────────────────────

// AnimePageResponse is the top-level response for the /anime endpoint.
type AnimePageResponse struct {
	Anime []ATAnime `json:"anime"`
	Links PageLinks `json:"links"`
	Meta  PageMeta  `json:"meta"`
}

// SongPageResponse is the top-level response for the /song endpoint.
type SongPageResponse struct {
	Songs []ATSongListItem `json:"songs"`
	Links PageLinks        `json:"links"`
	Meta  PageMeta         `json:"meta"`
}

// ThemeShowResponse wraps a single animetheme resource.
type ThemeShowResponse struct {
	AnimeTheme ATTheme `json:"animetheme"`
}

// AnimeShowResponse wraps a single anime resource.
type AnimeShowResponse struct {
	Anime ATAnime `json:"anime"`
}

// PageLinks holds first/last/prev/next pagination URLs.
type PageLinks struct {
	First *string `json:"first"`
	Last  *string `json:"last"`
	Prev  *string `json:"prev"`
	Next  *string `json:"next"`
}

// PageMeta holds pagination metadata.
type PageMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	From        int `json:"from"`
	To          int `json:"to"`
}

// ATAnime represents a single anime entry from AnimeThemes.
type ATAnime struct {
	ID          uint64       `json:"id"`
	Name        string       `json:"name"`
	Slug        string       `json:"slug"`
	Year        int          `json:"year"`
	Season      string       `json:"season"`       // "Winter","Spring","Summer","Fall"
	MediaFormat string       `json:"media_format"` // "TV","OVA","Movie","ONA","Special","Music"
	Synopsis    *string      `json:"synopsis"`
	Resources   []ATResource `json:"resources"`
	AnimeThemes []ATTheme    `json:"animethemes"`
}

// ATResource holds external links (including anilist_id).
type ATResource struct {
	ID         uint64  `json:"id"`
	Link       *string `json:"link"`
	ExternalID *int    `json:"external_id"`
	Site       string  `json:"site"` // "AniList","MyAnimeList","AniDB", etc.
}

// ATTheme is an OP/ED entry linked to an anime.
type ATTheme struct {
	ID       uint64    `json:"id"`
	Type     string    `json:"type"`     // "OP" | "ED"
	Sequence *int      `json:"sequence"` // nil = OP1/ED1
	Slug     string    `json:"slug"`
	Anime    *ATAnime  `json:"anime,omitempty"`
	Song     *ATSong   `json:"song"`
	Entries  []ATEntry `json:"animethemeentries"`
}

// ATSongListItem is a song row from the /song index (may include theme stubs).
type ATSongListItem struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	AnimeThemes []ATTheme `json:"animethemes"`
}

// ATSong is the song metadata of a theme.
type ATSong struct {
	ID      uint64     `json:"id"`
	Title   string     `json:"title"`
	Artists []ATArtist `json:"artists"`
}

// ATArtist represents a performing artist.
type ATArtist struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	// AnimeThemes returns anilist_id via resources on the artist endpoint,
	// but on the anime include it's not present — we store anime_themes_id only.
}

// ATEntry is a video version (variant) of a theme.
type ATEntry struct {
	ID       uint64    `json:"id"`
	Version  *int      `json:"version"` // nil = version 1
	Spoiler  bool      `json:"spoiler"`
	NSFW     bool      `json:"nsfw"`
	Episodes *string   `json:"episodes"`
	Videos   []ATVideo `json:"videos"`
}

// ATVideo holds the actual video file reference.
type ATVideo struct {
	ID       uint64 `json:"id"`
	Basename string `json:"basename"` // e.g. "ChainsawMan-OP1.webm"
	Filename string `json:"filename"` // e.g. "ChainsawMan-OP1"
	Path     string `json:"path"`     // e.g. "2022/fall/ChainsawMan-OP1.webm"
	Size     int64  `json:"size"`
	Tags     string `json:"tags"` // e.g. "1080", "NC", etc.
}

// ─── Client Methods ───────────────────────────────────────────────────────────

// FetchAnimePage fetches a paginated page of animes including themes, songs,
// artists, entries, videos, and AniList resources.
func (c *Client) FetchAnimePage(ctx context.Context, page, pageSize int) (*AnimePageResponse, error) {
	url := fmt.Sprintf(
		"%s/anime?include=animethemes.song.artists,animethemes.animethemeentries.videos,resources&page[size]=%d&page[number]=%d",
		baseURL, pageSize, page,
	)
	var result AnimePageResponse
	if err := c.getJSON(ctx, url, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FetchSongPage fetches songs with id greater than idGt (watermark), ascending by id.
// Includes theme stubs (ids only) — use FetchThemeByID for full graph data.
func (c *Client) FetchSongPage(ctx context.Context, page, pageSize int, idGt uint64) (*SongPageResponse, error) {
	u := fmt.Sprintf(
		"%s/song?include=animethemes&page[size]=%d&page[number]=%d&sort=id&filter[id-gt]=%d",
		baseURL, pageSize, page, idGt,
	)
	var result SongPageResponse
	if err := c.getJSON(ctx, u, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// FetchThemeByID loads a single animetheme with anime, song artists, entries, and videos.
// Uses the show endpoint because filter[id]+include often returns empty nested relations.
func (c *Client) FetchThemeByID(ctx context.Context, themeID uint64) (*ATTheme, error) {
	u := fmt.Sprintf(
		"%s/animetheme/%d?include=anime,song.artists,animethemeentries.videos",
		baseURL, themeID,
	)
	var result ThemeShowResponse
	if err := c.getJSON(ctx, u, &result); err != nil {
		return nil, err
	}
	if result.AnimeTheme.ID == 0 {
		return nil, fmt.Errorf("animethemes: theme %d not found", themeID)
	}
	return &result.AnimeTheme, nil
}

// FetchAnimeBySlug loads anime resources (AniList id, etc.) by slug.
func (c *Client) FetchAnimeBySlug(ctx context.Context, slug string) (*ATAnime, error) {
	u := fmt.Sprintf("%s/anime/%s?include=resources", baseURL, url.PathEscape(slug))
	var result AnimeShowResponse
	if err := c.getJSON(ctx, u, &result); err != nil {
		return nil, err
	}
	if result.Anime.ID == 0 {
		return nil, fmt.Errorf("animethemes: anime slug %q not found", slug)
	}
	return &result.Anime, nil
}

func (c *Client) getJSON(ctx context.Context, rawURL string, dest interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("animethemes: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AniRank/1.0 (https://anirank.app)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("animethemes: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("animethemes: rate limited (429)")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("animethemes: not found")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("animethemes: unexpected status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("animethemes: decode response: %w", err)
	}
	return nil
}
