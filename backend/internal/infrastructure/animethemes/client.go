package animethemes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseURL = "https://api.animethemes.moe"

// AnimeThemesClient defines the interface for the AnimeThemes API.
type AnimeThemesClient interface {
	FetchAnimePage(ctx context.Context, page, pageSize int) (*AnimePageResponse, error)
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
	ID          uint64      `json:"id"`
	Name        string      `json:"name"`
	Slug        string      `json:"slug"`
	Year        int         `json:"year"`
	Season      string      `json:"season"`       // "Winter","Spring","Summer","Fall"
	MediaFormat string      `json:"media_format"` // "TV","OVA","Movie","ONA","Special","Music"
	Synopsis    *string     `json:"synopsis"`
	Resources   []ATResource `json:"resources"`
	AnimeThemes []ATTheme   `json:"animethemes"`
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
	Song     *ATSong   `json:"song"`
	Entries  []ATEntry `json:"animethemeentries"`
}

// ATSong is the song metadata of a theme.
type ATSong struct {
	ID      uint64     `json:"id"`
	Title   string     `json:"title"`
	Artists []ATArtist `json:"artists"`
}

// ATArtist represents a performing artist.
type ATArtist struct {
	ID   uint64  `json:"id"`
	Name string  `json:"name"`
	Slug string  `json:"slug"`
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

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("animethemes: build request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "AniRank/1.0 (https://anirank.app)")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("animethemes: do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("animethemes: rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("animethemes: unexpected status %d", resp.StatusCode)
	}

	var result AnimePageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("animethemes: decode response: %w", err)
	}

	return &result, nil
}
