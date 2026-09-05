package rate

import (
	"errors"
	"strings"
	"testing"

	"anirank/api/internal/domain"
)

type stubURLResolver struct{}

func (stubURLResolver) GetURL(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(path), "http") {
		return path
	}
	return "https://cdn.test/" + path
}

func TestValidateSongForRate(t *testing.T) {
	activeAnime := &domain.Anime{Status: true}
	inactiveAnime := &domain.Anime{Status: false}

	tests := []struct {
		name    string
		song    *domain.Song
		audio   string
		wantErr error
	}{
		{name: "nil song", song: nil, wantErr: domain.ErrNotFound},
		{
			name:    "inactive song",
			song:    &domain.Song{Status: false, Anime: activeAnime},
			audio:   "https://cdn.test/x",
			wantErr: ErrRateSongInactive,
		},
		{
			name:    "missing anime",
			song:    &domain.Song{Status: true},
			audio:   "https://cdn.test/x",
			wantErr: ErrRateAnimeInactive,
		},
		{
			name:    "inactive anime",
			song:    &domain.Song{Status: true, Anime: inactiveAnime},
			audio:   "https://cdn.test/x",
			wantErr: ErrRateAnimeInactive,
		},
		{
			name:    "no media",
			song:    &domain.Song{Status: true, Anime: activeAnime},
			audio:   "",
			wantErr: ErrRateSongNoMedia,
		},
		{
			name:    "ok",
			song:    &domain.Song{Status: true, Anime: activeAnime},
			audio:   "https://cdn.test/x",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSongForRate(tt.song, tt.audio)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRateSongLoadErrorMessage(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{ErrRateSongInactive, "This theme is inactive or not publicly available"},
		{ErrRateAnimeInactive, "This theme is inactive or not publicly available"},
		{ErrRateSongNoMedia, "This theme has no playable video"},
		{domain.ErrNotFound, "Song not found"},
		{errors.New("db down"), "Song not found"},
		{nil, "Song not found"},
	}
	for _, c := range cases {
		if got := rateSongLoadErrorMessage(c.err); got != c.want {
			t.Errorf("rateSongLoadErrorMessage(%v)=%q want %q", c.err, got, c.want)
		}
	}
}

func TestResolvePlayableAudioURLPrefersActiveVariant(t *testing.T) {
	activeSrc := "active.webm"
	inactiveSrc := "inactive.webm"
	song := &domain.Song{
		Variants: []domain.SongVariant{
			{
				Status: false,
				Video: &domain.SongVariantVideo{
					VideoSrc: &inactiveSrc,
					LocalUrl: &inactiveSrc,
				},
			},
			{
				Status: true,
				Video: &domain.SongVariantVideo{
					VideoSrc: &activeSrc,
					LocalUrl: &activeSrc,
				},
			},
		},
	}

	got := resolvePlayableAudioURL(stubURLResolver{}, song)
	want := "https://cdn.test/active.webm"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolvePlayableAudioURLFallbackInactive(t *testing.T) {
	src := "only-inactive.webm"
	song := &domain.Song{
		Variants: []domain.SongVariant{
			{
				Status: false,
				Video: &domain.SongVariantVideo{
					VideoSrc: &src,
					LocalUrl: &src,
				},
			},
		},
	}
	got := resolvePlayableAudioURL(stubURLResolver{}, song)
	want := "https://cdn.test/only-inactive.webm"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolvePlayableAudioURLEmpty(t *testing.T) {
	song := &domain.Song{
		Variants: []domain.SongVariant{{Status: true}},
	}
	if got := resolvePlayableAudioURL(stubURLResolver{}, song); got != "" {
		t.Fatalf("expected empty, got %q", got)
	}
}
