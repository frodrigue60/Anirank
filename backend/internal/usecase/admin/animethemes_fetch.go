package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	atAPIBaseURL      = "https://api.animethemes.moe"
	atIncludeParams   = "animethemes.song.artists,animethemes.group,images,animethemes.animethemeentries.videos,studios,resources"
	atSeasonPageSize  = 100
	atInterFetchDelay = 250 * time.Millisecond
)

type atPaginatedResponse struct {
	Anime []ATAnimeData `json:"anime"`
	Links struct {
		Next *string `json:"next"`
	} `json:"links"`
}

func newATHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func atUserAgentRequest(ctx context.Context, method, rawURL string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Anirank/1.0 (https://anirank.work)")
	return req, nil
}

// fetchSeasonAnimeStubs returns all anime stubs for a season, paginating through the AT list endpoint.
func fetchSeasonAnimeStubs(ctx context.Context, client *http.Client, year int, seasonName string) ([]ATAnimeData, error) {
	var all []ATAnimeData
	page := 1

	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		rawURL := fmt.Sprintf(
			"%s/anime?include=%s&filter[year]=%d&filter[season]=%s&page[size]=%d&page[number]=%d",
			atAPIBaseURL, atIncludeParams, year, strings.ToLower(seasonName), atSeasonPageSize, page,
		)

		req, err := atUserAgentRequest(ctx, http.MethodGet, rawURL)
		if err != nil {
			return nil, fmt.Errorf("build season list request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("fetch season page %d: %w", page, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("AnimeThemes API returned status %d on page %d", resp.StatusCode, page)
		}

		var atResp atPaginatedResponse
		if err := json.NewDecoder(resp.Body).Decode(&atResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode season page %d: %w", page, err)
		}
		resp.Body.Close()

		all = append(all, atResp.Anime...)

		if atResp.Links.Next == nil || *atResp.Links.Next == "" {
			break
		}

		page++
		time.Sleep(atInterFetchDelay)
	}

	return all, nil
}

func deepFetchAnimeBySlug(ctx context.Context, client *http.Client, slug string) (*ATAnimeData, error) {
	rawURL := fmt.Sprintf("%s/anime/%s?include=%s", atAPIBaseURL, slug, atIncludeParams)

	req, err := atUserAgentRequest(ctx, http.MethodGet, rawURL)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AnimeThemes detail status %d for slug %s", resp.StatusCode, slug)
	}

	var detailResp struct {
		Anime ATAnimeData `json:"anime"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
		return nil, err
	}

	return &detailResp.Anime, nil
}

// deepFetchAnimeCollection deep-fetches each anime by slug so nested songs, entries, and videos are populated.
// AnimeThemes list/filter endpoints are unreliable for nested includes.
func deepFetchAnimeCollection(ctx context.Context, client *http.Client, stubs []ATAnimeData, progress chan<- string) []ATAnimeData {
	sendProgress := func(msg string) {
		if progress != nil {
			select {
			case progress <- msg:
			default:
			}
		}
	}

	fullList := make([]ATAnimeData, 0, len(stubs))
	total := len(stubs)

	for i, stub := range stubs {
		if ctx.Err() != nil {
			break
		}

		sendProgress(fmt.Sprintf("[%d/%d] Fetching full record for: %s", i+1, total, stub.Name))

		anime, err := deepFetchAnimeBySlug(ctx, client, stub.Slug)
		if err != nil {
			log.Printf("[ERROR] Failed to deep-fetch slug %s: %v\n", stub.Slug, err)
			sendProgress(fmt.Sprintf("Warning: could not fetch %s — skipping", stub.Name))
			if i < total-1 {
				time.Sleep(atInterFetchDelay)
			}
			continue
		}

		fullList = append(fullList, *anime)

		if i < total-1 {
			time.Sleep(atInterFetchDelay)
		}
	}

	return fullList
}
