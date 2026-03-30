package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const AnilistGraphQLEndpoint = "https://graphql.anilist.co"
const AnilistTokenEndpoint = "https://anilist.co/api/v2/oauth/token"
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/123.0.0.0 Safari/537.36"

func (c *Client) setAdvancedHeaders(req *http.Request) {
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://anilist.co")
	req.Header.Set("Referer", "https://anilist.co/")
	req.Header.Set("Sec-Ch-Ua", ` "Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123" `)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
}

type Client struct {
	httpClient *http.Client
}

type AnilistUser struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GraphQLQuery is the structure for the request body
type GraphQLQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// AnilistResponse represents the structure of the JSON response from Anilist
type AnilistResponse struct {
	Data struct {
		Page struct {
			PageInfo struct {
				Total       int  `json:"total"`
				PerPage     int  `json:"perPage"`
				CurrentPage int  `json:"currentPage"`
				LastPage    int  `json:"lastPage"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
			Media []Media `json:"media"`
		} `json:"Page"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type Media struct {
	ID          int    `json:"id"`
	Title       Title  `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CoverImage  struct {
		ExtraLarge string `json:"extraLarge"`
		Large      string `json:"large"`
	} `json:"coverImage"`
	BannerImage string   `json:"bannerImage"`
	Season      string   `json:"season"`
	SeasonYear  int      `json:"seasonYear"`
	Format      string   `json:"format"`
	Genres      []string `json:"genres"`
	Studios     struct {
		Edges []struct {
			IsMain bool `json:"isMain"`
			Node   struct {
				Name string `json:"name"`
			} `json:"node"`
		} `json:"edges"`
	} `json:"studios"`
	ExternalLinks []struct {
		Site string `json:"site"`
		URL  string `json:"url"`
	} `json:"externalLinks"`
}

type Staff struct {
	ID   int `json:"id"`
	Name struct {
		Full        string   `json:"full"`
		Native      string   `json:"native"`
		Alternative []string `json:"alternative"`
	} `json:"name"`
	Image struct {
		Large string `json:"large"`
	} `json:"image"`
}

type StaffResponse struct {
	Data struct {
		Page struct {
			Staff []Staff `json:"staff"`
		} `json:"Page"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

type Title struct {
	Romaji  string `json:"romaji"`
	English string `json:"english"`
	Native  string `json:"native"`
}

// the query to search animes by title
const searchMediaQuery = `
query ($page: Int, $perPage: Int, $search: String, $format: MediaFormat) {
	Page(page: $page, perPage: $perPage) {
		pageInfo {
			total
			perPage
			currentPage
			lastPage
			hasNextPage
		}
		media(search: $search, type: ANIME, isAdult: false, format: $format, sort: SEARCH_MATCH) {
			id
			title {
				romaji
				english
				native
			}
			description(asHtml: false)
			status
			coverImage {
				extraLarge
				large
			}
			bannerImage
			season
			seasonYear
			format
			genres
			studios {
				edges {
					isMain
					node {
						name
					}
				}
			}
			externalLinks {
				site
				url
			}
		}
	}
}
`

const searchStaffQuery = `
query ($search: String) {
  Page(page: 1, perPage: 5) {
    staff(search: $search) {
      id
      name {
        full
        native
        alternative
      }
      image {
        large
      }
    }
  }
}
`

// SearchAnimes searches animes on Anilist by title
func (c *Client) SearchAnimes(ctx context.Context, search string, format string, page int) (*AnilistResponse, error) {
	variables := map[string]interface{}{
		"page":    page,
		"perPage": 20,
		"search":  search,
	}

	if format != "" {
		variables["format"] = format
	}

	query := GraphQLQuery{
		Query:     searchMediaQuery,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist api returned status: %d", resp.StatusCode)
	}

	var anilistResp AnilistResponse
	if err := json.NewDecoder(resp.Body).Decode(&anilistResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anilistResp.Errors) > 0 {
		return nil, fmt.Errorf("anilist api error: %s", anilistResp.Errors[0].Message)
	}

	return &anilistResp, nil
}

// multiMediaByIDQuery fetches multiple media items by their IDs
const multiMediaByIDQuery = `
query ($ids: [Int]) {
	Page(page: 1, perPage: 50) {
		media(id_in: $ids, type: ANIME, isAdult: false) {
			id
			title {
				romaji
				english
				native
			}
			description(asHtml: false)
			status
			coverImage {
				extraLarge
				large
			}
			bannerImage
			season
			seasonYear
			format
			genres
			studios {
				edges {
					isMain
					node {
						name
					}
				}
			}
			externalLinks {
				site
				url
			}
		}
	}
}
`

// GetMediaByIDs fetches multiple media items from Anilist by their IDs
func (c *Client) GetMediaByIDs(ctx context.Context, ids []int) ([]Media, error) {
	variables := map[string]interface{}{
		"ids": ids,
	}

	query := GraphQLQuery{
		Query:     multiMediaByIDQuery,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist api returned status: %d", resp.StatusCode)
	}

	var anilistResp AnilistResponse
	if err := json.NewDecoder(resp.Body).Decode(&anilistResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anilistResp.Errors) > 0 {
		return nil, fmt.Errorf("anilist api error: %s", anilistResp.Errors[0].Message)
	}

	return anilistResp.Data.Page.Media, nil
}

// the query to fetch season animes
const seasonMediaQuery = `
query ($page: Int, $perPage: Int, $season: MediaSeason, $seasonYear: Int, $format: MediaFormat) {
	Page(page: $page, perPage: $perPage) {
		pageInfo {
			total
			perPage
			currentPage
			lastPage
			hasNextPage
		}
		media(season: $season, seasonYear: $seasonYear, format: $format, type: ANIME, isAdult: false, sort: POPULARITY_DESC) {
			id
			title {
				romaji
				english
				native
			}
			description(asHtml: false)
			status
			coverImage {
				extraLarge
				large
			}
			bannerImage
			season
			seasonYear
			format
			genres
			studios {
				edges {
					isMain
					node {
						name
					}
				}
			}
			externalLinks {
				site
				url
			}
		}
	}
}
`

// FetchAnimes fetches a page of animes from Anilist
func (c *Client) FetchAnimes(ctx context.Context, page int, season string, seasonYear int, format string) (*AnilistResponse, error) {
	variables := map[string]interface{}{
		"page":       page,
		"perPage":    50,
		"season":     season,
		"seasonYear": seasonYear,
	}

	if format != "" {
		variables["format"] = format
	}

	query := GraphQLQuery{
		Query:     seasonMediaQuery,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist api returned status: %d", resp.StatusCode)
	}

	var anilistResp AnilistResponse
	if err := json.NewDecoder(resp.Body).Decode(&anilistResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(anilistResp.Errors) > 0 {
		return nil, fmt.Errorf("anilist api error: %s", anilistResp.Errors[0].Message)
	}

	return &anilistResp, nil
}

// ExchangeCode exchanges the authorization code for an access token
func (c *Client) ExchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("redirect_uri", redirectURI)
	data.Set("code", code)

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistTokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			Message          string `json:"message"`
			ErrorDescription string `json:"error_description"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Message != "" {
			return nil, fmt.Errorf("anilist oauth: %s (%s)", errResp.Error, errResp.Message)
		}
		if errResp.ErrorDescription != "" {
			return nil, fmt.Errorf("anilist oauth: %s: %s", errResp.Error, errResp.ErrorDescription)
		}
		if errResp.Error != "" {
			return nil, fmt.Errorf("anilist oauth: %s", errResp.Error)
		}
		return nil, fmt.Errorf("anilist token exchange failed with HTTP %d", resp.StatusCode)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

// GetViewer fetches the authenticated user's profile info
func (c *Client) GetViewer(ctx context.Context, accessToken string) (*AnilistUser, error) {
	query := `query { Viewer { id name } }`
	payload := GraphQLQuery{
		Query: query,
	}

	bodyBytes, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Viewer AnilistUser `json:"Viewer"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.Data.Viewer, nil
}

// SearchStaff searches staff on Anilist by name
func (c *Client) SearchStaff(ctx context.Context, search string) ([]Staff, error) {
	variables := map[string]interface{}{
		"search": search,
	}

	query := GraphQLQuery{
		Query:     searchStaffQuery,
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded (429)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist api returned status: %d", resp.StatusCode)
	}

	var staffResp StaffResponse
	if err := json.NewDecoder(resp.Body).Decode(&staffResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(staffResp.Errors) > 0 {
		return nil, fmt.Errorf("anilist api error: %s", staffResp.Errors[0].Message)
	}

	return staffResp.Data.Page.Staff, nil
}

type StaffSearchReq struct {
	ID   *uint64
	Name string
}

// SearchStaffBatch searches multiple staff on Anilist by names or IDs using GraphQL aliases
func (c *Client) SearchStaffBatch(ctx context.Context, reqs []StaffSearchReq) (map[string][]Staff, error) {
	if len(reqs) == 0 {
		return make(map[string][]Staff), nil
	}

	// 1. Build the dynamic GraphQL query with aliases INSIDE a single Page query
	var queryBuilder strings.Builder
	queryBuilder.WriteString("query (")
	for i, req := range reqs {
		if i > 0 {
			queryBuilder.WriteString(", ")
		}
		if req.ID != nil {
			fmt.Fprintf(&queryBuilder, "$id%d: Int", i)
		} else {
			fmt.Fprintf(&queryBuilder, "$n%d: String", i)
		}
	}
	queryBuilder.WriteString(") {\n")

	for i, req := range reqs {
		if req.ID != nil {
			fmt.Fprintf(&queryBuilder, "  s%d: staff(id: $id%d) { id name { full native alternative } image { large } }\n", i, i)
		} else {
			fmt.Fprintf(&queryBuilder, "  s%d: Page(page: 1, perPage: 1) { staff(search: $n%d) { id name { full native alternative } image { large } } }\n", i, i)
		}
	}
	queryBuilder.WriteString("}")

	variables := make(map[string]interface{})
	for i, req := range reqs {
		if req.ID != nil {
			variables[fmt.Sprintf("id%d", i)] = *req.ID
		} else {
			variables[fmt.Sprintf("n%d", i)] = req.Name
		}
	}

	payload := GraphQLQuery{
		Query:     queryBuilder.String(),
		Variables: variables,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal batch query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", AnilistGraphQLEndpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create batch request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	c.setAdvancedHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("batch request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limit exceeded (429)")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anilist batch api returned status: %d", resp.StatusCode)
	}

	// 2. Parse the dynamic aliased response
	var result struct {
		Data map[string]interface{} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode batch response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("anilist batch api error: %s", result.Errors[0].Message)
	}

	// 3. Map back to original name keys
	finalResult := make(map[string][]Staff)
	data := result.Data

	for i, req := range reqs {
		alias := fmt.Sprintf("s%d", i)
		if val, ok := data[alias]; ok && val != nil {
			var staffs []Staff
			
			// Handle both s%d: staff (ID search) and s%d: Page { staff: [] } (Name search)
			sMap, isMap := val.(map[string]interface{})
			if isMap {
				if staffList, ok := sMap["staff"].([]interface{}); ok {
					// It's the Page { staff: [] } structure
					for _, s := range staffList {
						if sm, ok := s.(map[string]interface{}); ok {
							staffs = append(staffs, decodeStaffMap(sm))
						}
					}
				} else if _, isID := sMap["id"]; isID {
					// It's the root staff structure (ID search)
					staffs = append(staffs, decodeStaffMap(sMap))
				}
			}
			
			finalResult[req.Name] = staffs
		}
	}

	return finalResult, nil
}

func decodeStaffMap(sMap map[string]interface{}) Staff {
	var st Staff
	if id, ok := sMap["id"].(float64); ok {
		st.ID = int(id)
	}
	
	if nameData, ok := sMap["name"].(map[string]interface{}); ok {
		if full, ok := nameData["full"].(string); ok {
			st.Name.Full = full
		}
		if native, ok := nameData["native"].(string); ok {
			st.Name.Native = native
		}
		if alt, ok := nameData["alternative"].([]interface{}); ok {
			for _, a := range alt {
				if as, ok := a.(string); ok {
					st.Name.Alternative = append(st.Name.Alternative, as)
				}
			}
		}
	}

	if imgData, ok := sMap["image"].(map[string]interface{}); ok {
		if large, ok := imgData["large"].(string); ok {
			st.Image.Large = large
		}
	}
	
	return st
}

