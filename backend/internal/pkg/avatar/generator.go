package avatar

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// Result contains the avatar image data and metadata
type Result struct {
	Data        []byte
	Size        int64
	ContentType string
}

// Generate fetches an initials-based avatar from ui-avatars.com
func Generate(ctx context.Context, name string) (*Result, error) {
	seed := url.QueryEscape(name)
	// Using random background, white text, and 512px size
	apiUrl := fmt.Sprintf("https://ui-avatars.com/api/?name=%s&background=random&color=fff&size=512&format=png", seed)

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", apiUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch avatar: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("avatar service returned status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read avatar data: %w", err)
	}

	return &Result{
		Data:        data,
		Size:        int64(len(data)),
		ContentType: "image/png",
	}, nil
}
