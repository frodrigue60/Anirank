package og

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"anirank/api/internal/domain"
	"anirank/api/internal/pkg/imageutil"

	"github.com/disintegration/gift"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
)

//go:embed assets/fonts/*.ttf
var fontAssets embed.FS

type Generator struct {
	fontBlack []byte
	fontBold  []byte
	fontMedi  []byte
	fontRegu  []byte
	version   int
	cacheDir  string
	s3PublicURL string
	s3Endpoint  string
}

func NewGenerator(s3PublicURL, s3Endpoint string) *Generator {
	// Initialize font paths from embedded FS
	fBlack, _ := fontAssets.ReadFile("assets/fonts/Inter-Black.ttf")
	fBold, _  := fontAssets.ReadFile("assets/fonts/Inter-Bold.ttf")
	fMedi, _  := fontAssets.ReadFile("assets/fonts/Inter-Medium.ttf")
	fRegu, _  := fontAssets.ReadFile("assets/fonts/Inter-Regular.ttf")

	// Create cache directory if it doesn't exist
	// We use a relative path from the CWD or a standard location
	cacheDir := "storage/og_cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			fmt.Printf("[OG] Warning: Failed to create cache directory: %v\n", err)
		}
	}

	// Load version from file
	version := 1
	versionPath := filepath.Join(cacheDir, "version.txt")
	if data, err := os.ReadFile(versionPath); err == nil {
		fmt.Sscanf(string(data), "%d", &version)
	}

	return &Generator{
		fontBlack: fBlack,
		fontBold:  fBold,
		fontMedi:  fMedi,
		fontRegu:  fRegu,
		version:   version,
		cacheDir:  cacheDir,
		s3PublicURL: s3PublicURL,
		s3Endpoint:  s3Endpoint,
	}
}

func (g *Generator) loadFont(data []byte, size float64) (font.Face, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("font data is empty")
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}
	return truetype.NewFace(f, &truetype.Options{
		Size: size,
	}), nil
}

func (g *Generator) GetVersion() int {
	return g.version
}

func (g *Generator) GetCache(key string) ([]byte, bool) {
	path := filepath.Join(g.cacheDir, key+".png")
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err == nil {
			return data, true
		}
	}
	return nil, false
}

func (g *Generator) SaveCache(key string, data []byte) error {
	path := filepath.Join(g.cacheDir, key+".png")
	return os.WriteFile(path, data, 0644)
}

func (g *Generator) FlushCache() error {
	entries, err := os.ReadDir(g.cacheDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".png") {
			path := filepath.Join(g.cacheDir, entry.Name())
			if err := os.Remove(path); err != nil {
				fmt.Printf("[OG] Failed to remove cache file %s: %v\n", path, err)
			}
		}
	}

	// Increment version and save
	g.version++
	versionPath := filepath.Join(g.cacheDir, "version.txt")
	os.WriteFile(versionPath, []byte(fmt.Sprintf("%d", g.version)), 0644)
	fmt.Printf("[OG] Cache flushed. New version: %d\n", g.version)

	return nil
}

func (g *Generator) GenerateSongOG(title, artists, animeTitle, songType string, score float64, bgUrl string) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)

	// 1. Background
	if bgUrl != "" {
		if err := g.drawBlurredBackground(dc, bgUrl, W, H); err != nil {
			fmt.Printf("[OG] Error drawing blurred background: %v\n", err)
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// 2. Artistic accents (optional)
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.1) // Primary color accent
	dc.DrawCircle(W, 0, 400)
	dc.Fill()

	// 3. Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", W-80, 60, 1, 0.5)
	}

	// 4. Song Type Badge
	if face, err := g.loadFont(g.fontBlack, 32); err == nil {
		dc.SetFontFace(face)
		dc.SetHexColor("#ff4e50")
		dc.DrawStringAnchored(strings.ToUpper(songType), 80, 110, 0, 0.5)
	}

	// 5. Song Title
	title = g.truncate(title, 55)
	if face, err := g.loadFont(g.fontBlack, 90); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringWrapped(title, 80, 160, 0, 0, 1000, 1.1, gg.AlignLeft)
	}

	// 6. Artist
	if artists != "" {
		artists = g.truncate(artists, 70)
		if face, err := g.loadFont(g.fontBold, 36); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.7)
			dc.DrawStringAnchored(artists, 80, 380, 0, 0.5)
		}
	}

	// 7. Info Bar (Bottom)
	if animeTitle != "" {
		animeTitle = g.truncate(animeTitle, 60)
		if face, err := g.loadFont(g.fontBold, 20); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.5)
			dc.DrawStringAnchored("FEATURED IN", 80, 500, 0, 0.5)
		}
		if face, err := g.loadFont(g.fontBlack, 42); err == nil {
			dc.SetFontFace(face)
			dc.SetRGB(1, 1, 1)
			dc.DrawStringAnchored(animeTitle, 80, 550, 0, 0.5)
		}
	}

	// Score
	if score > 0 {
		if face, err := g.loadFont(g.fontBold, 20); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.5)
			dc.DrawStringAnchored("COMMUNITY SCORE", W-80, 500, 1, 0.5)
		}
		if face, err := g.loadFont(g.fontBlack, 64); err == nil {
			dc.SetFontFace(face)
			dc.SetHexColor("#FFD700")
			dc.DrawStringAnchored(fmt.Sprintf("★ %.1f%%", score), W-80, 550, 1, 0.5)
		}
	}

	return dc.Image(), nil
}

func (g *Generator) GenerateArtistOG(name string, songCount int, favoriteCount int, avatarUrl, bannerUrl string) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)

	// 1. Background (Blurred banner or avatar)
	bgUrl := bannerUrl
	if bgUrl == "" {
		bgUrl = avatarUrl
	}

	if bgUrl != "" {
		if err := g.drawBlurredBackground(dc, bgUrl, W, H); err != nil {
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// 2. Artistic accents
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.15)
	dc.DrawCircle(W, 0, 400)
	dc.Fill()

	// 3. Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", W-80, 60, 1, 0.5)
	}

	// 4. Centered Avatar Circle
	if avatarUrl != "" {
		img, err := g.fetchImage(avatarUrl)
		if err == nil {
			radius := 140.0
			avatarX := W / 2.0
			avatarY := H/2.0 - 80.0
			diameter := int(radius * 2)

			// Resize avatar to fit circle
			resizer := gift.New(
				gift.ResizeToFill(diameter, diameter, gift.LanczosResampling, gift.CenterAnchor),
			)
			resized := image.NewRGBA(resizer.Bounds(img.Bounds()))
			resizer.Draw(resized, img)

			// Image
			dc.DrawCircle(avatarX, avatarY, radius)
			dc.Clip()
			dc.DrawImageAnchored(resized, int(avatarX), int(avatarY), 0.5, 0.5)
			dc.ResetClip()
		}
	}

	// 5. Name
	name = g.truncate(name, 40)
	if face, err := g.loadFont(g.fontBlack, 84); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringAnchored(name, W/2, H-200, 0.5, 0.5)
	}

	// 6. Stats
	if face, err := g.loadFont(g.fontBold, 34); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.4)
		statsText := fmt.Sprintf("%d SONGS • %d FAVORITES", songCount, favoriteCount)
		dc.DrawStringAnchored(statsText, W/2, H-120, 0.5, 0.5)
	}

	// Branding URL
	if face, err := g.loadFont(g.fontBold, 18); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.3)
		dc.DrawStringAnchored("ANIRANK.WORK", W/2, H-40, 0.5, 0.5)
	}

	return dc.Image(), nil
}

func (g *Generator) GeneratePlaylistOG(name, creator string, songCount int, bannerUrl string) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)

	// 1. Background (Blurred banner)
	if bannerUrl != "" {
		if err := g.drawBlurredBackground(dc, bannerUrl, W, H); err != nil {
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// 2. Artistic accents
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.2)
	dc.DrawCircle(0, 0, 400)
	dc.Fill()

	// 3. Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", W-80, 60, 1, 0.5)
	}

	// 4. Playlist Icon Badge
	if face, err := g.loadFont(g.fontBlack, 32); err == nil {
		dc.SetFontFace(face)
		dc.SetHexColor("#7f13ec")
		dc.DrawStringAnchored("PLAYLIST", W/2, H/2-180, 0.5, 0.5)
	}

	// 5. Name
	name = g.truncate(name, 50)
	if face, err := g.loadFont(g.fontBlack, 90); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringAnchored(name, W/2, H/2-80, 0.5, 0.5)
	}

	// 6. Creator
	creatorText := "Featured Playlist"
	if creator != "" {
		creatorText = fmt.Sprintf("Curated by %s", creator)
	}
	if face, err := g.loadFont(g.fontBold, 36); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.7)
		dc.DrawStringAnchored(creatorText, W/2, H/2+20, 0.5, 0.5)
	}

	// 7. Stats
	if face, err := g.loadFont(g.fontBold, 34); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.5)
		dc.DrawStringAnchored(fmt.Sprintf("%d SONGS", songCount), W/2, H/2+120, 0.5, 0.5)
	}

	// Branding URL
	if face, err := g.loadFont(g.fontBold, 18); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.3)
		dc.DrawStringAnchored("ANIRANK.WORK", W/2, H-40, 0.5, 0.5)
	}

	return dc.Image(), nil
}

func (g *Generator) GenerateUserOG(name string, level int, xp int, followers, ratings int, avatarUrl, bannerUrl string) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)

	// 1. Background (Banner or blurred avatar)
	bgUrl := bannerUrl
	if bgUrl == "" {
		bgUrl = avatarUrl
	}

	fmt.Printf("[OG] User=%s, bgUrl=%s, avatarUrl=%s\n", name, bgUrl, avatarUrl)

	if bgUrl != "" {
		if err := g.drawBlurredBackground(dc, bgUrl, W, H); err != nil {
			fmt.Printf("[OG] Failed to draw background for %s: %v\n", name, err)
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// 2. Stats Bar Bottom (Anilist Style)
	barHeight := 180.0
	dc.SetRGBA(15.0/255.0, 23.0/255.0, 42.0/255.0, 0.85) // Dark Navy overlay
	dc.DrawRectangle(0, H-barHeight, W, barHeight)
	dc.Fill()

	// 3. Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", W-80, 60, 1, 0.5)
	}

	// 4. User Avatar (Overlapping)
	if avatarUrl != "" {
		img, err := g.fetchImage(avatarUrl)
		if err == nil {
			radius := 140.0
			avatarX := 200.0
			avatarY := H - barHeight
			diameter := int(radius * 2)

			// Resize avatar to fit circle
			resizer := gift.New(
				gift.ResizeToFill(diameter, diameter, gift.LanczosResampling, gift.CenterAnchor),
			)
			resized := image.NewRGBA(resizer.Bounds(img.Bounds()))
			resizer.Draw(resized, img)

			// Border
			dc.SetHexColor("#7f13ec")
			dc.DrawCircle(avatarX, avatarY, radius+8)
			dc.Fill()

			// Image
			dc.DrawCircle(avatarX, avatarY, radius)
			dc.Clip()
			dc.DrawImageAnchored(resized, int(avatarX), int(avatarY), 0.5, 0.5)
			dc.ResetClip()
		}
	}

	// 5. Name
	name = g.truncate(name, 30)
	if face, err := g.loadFont(g.fontBlack, 80); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringAnchored(name, 380, H-barHeight-40, 0, 0.5)
	}

	// 6. Stats Row
	statsY := H - (barHeight / 2)
	statsStartX := 400.0
	spacing := 280.0

	drawStat := func(x float64, label string, value string) {
		if face, err := g.loadFont(g.fontBlack, 64); err == nil {
			dc.SetFontFace(face)
			dc.SetHexColor("#7f13ec")
			dc.DrawStringAnchored(value, x, statsY-10, 0.5, 0.5)
		}

		if face, err := g.loadFont(g.fontBold, 24); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.6)
			dc.DrawStringAnchored(strings.ToUpper(label), x, statsY+40, 0.5, 0.5)
		}
	}

	drawStat(statsStartX, "Level", fmt.Sprintf("%d", level))
	drawStat(statsStartX+spacing, "Followers", fmt.Sprintf("%d", followers))
	drawStat(statsStartX+spacing*2, "Ratings", fmt.Sprintf("%d", ratings))

	return dc.Image(), nil
}

func (g *Generator) GenerateAnimeOG(title, studios string, songCount int, score float64, bgUrl string) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)
	drawGradientBackground(dc, W, H)

	if bgUrl != "" {
		if err := g.drawBlurredBackground(dc, bgUrl, W, H); err != nil {
			fmt.Printf("[OG] Error drawing blurred background: %v\n", err)
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", W-80, 60, 1, 0.5)
	}

	// Title
	title = g.truncate(title, 60)
	if face, err := g.loadFont(g.fontBlack, 80); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		// Start higher and align top
		dc.DrawStringWrapped(title, 80, 130, 0, 0, 1000, 1.1, gg.AlignLeft)
	}

	// Studios
	if studios != "" {
		studios = g.truncate(studios, 80)
		if face, err := g.loadFont(g.fontBold, 34); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.7)
			// Positioned much lower to avoid overlap with 3-line title
			dc.DrawStringAnchored(strings.ToUpper(studios), 80, 440, 0, 0.5)
		}
	}

	// Score
	if score > 0 {
		if face, err := g.loadFont(g.fontBlack, 64); err == nil {
			dc.SetFontFace(face)
			dc.SetHexColor("#FFD700")
			dc.DrawStringAnchored(fmt.Sprintf("★ %.1f", score), 80, 510, 0, 0.5)
		}
	}

	// Bottom Text (Song Count)
	bottomText := "Discover • Rate • Rank"
	if songCount > 0 {
		bottomText = fmt.Sprintf("%d Songs Available", songCount)
	}

	if face, err := g.loadFont(g.fontBold, 28); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.4)
		dc.DrawStringAnchored(bottomText, 80, 580, 0, 0.5)
	}

	return dc.Image(), nil
}

func (g *Generator) GenerateHomeOG(totalSongs, totalUsers, totalAnimes, totalArtists int) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)
	drawGradientBackground(dc, W, H)

	// Artistic accents
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.2)
	dc.DrawCircle(0, 0, 500)
	dc.Fill()
	dc.DrawCircle(W, H, 400)
	dc.Fill()

	// Title
	if face, err := g.loadFont(g.fontBlack, 130); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringAnchored("ANIRANK", W/2, H/2-250, 0.5, 0.5)
	}

	// Tagline
	if face, err := g.loadFont(g.fontBold, 34); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("The Ultimate Anime Music Ranking Platform", W/2, H/2-170, 0.5, 0.5)
	}

	// Stats 2x2 Grid
	statsY := H/2 - 50.0
	col1X := W/2 - 180.0
	col2X := W/2 + 180.0
	rowHeight := 100.0

	drawStatItem := func(x, y float64, label string, value int) {
		// Value
		if face, err := g.loadFont(g.fontBlack, 54); err == nil {
			dc.SetFontFace(face)
			dc.SetRGB(1, 1, 1)
			dc.DrawStringAnchored(fmt.Sprintf("%d", value), x, y, 0.5, 0.5)
		}
		// Label
		if face, err := g.loadFont(g.fontBold, 22); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.6)
			dc.DrawStringAnchored(strings.ToUpper(label), x, y+40, 0.5, 0.5)
		}
	}

	drawStatItem(col1X, statsY, "Animes", totalAnimes)
	drawStatItem(col2X, statsY, "Songs", totalSongs)
	drawStatItem(col1X, statsY+rowHeight, "Artists", totalArtists)
	drawStatItem(col2X, statsY+rowHeight, "Users", totalUsers)

	// Sub-tagline
	if face, err := g.loadFont(g.fontBold, 24); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.4)
		dc.DrawStringAnchored("Discover • Rate • Share", W/2, H-120, 0.5, 0.5)
	}

	// Subtle accent line
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.5)
	dc.DrawRectangle(W/2-60, H-90, 120, 4)
	dc.Fill()

	// Branding URL
	if face, err := g.loadFont(g.fontBold, 22); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.6)
		dc.DrawStringAnchored("ANIRANK.WORK", W/2, H-50, 0.5, 0.5)
	}

	return dc.Image(), nil
}

func (g *Generator) drawBlurredBackground(dc *gg.Context, url string, w, h float64) error {
	img, err := g.fetchImage(url)
	if err != nil {
		return err
	}

	// 1. Process image with GIFT
	gi := gift.New(
		gift.ResizeToFill(int(w), int(h), gift.LanczosResampling, gift.CenterAnchor),
		gift.GaussianBlur(8),
	)
	
	dst := image.NewRGBA(gi.Bounds(img.Bounds()))
	gi.Draw(dst, img)

	// 2. Draw to GG
	dc.DrawImage(dst, 0, 0)

	// 3. Overlay dark mask for readability
	dc.SetRGBA(0, 0, 0, 0.7)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()

	return nil
}

func drawGradientBackground(dc *gg.Context, w, h float64) {
	grad := gg.NewLinearGradient(0, 0, w, h)
	grad.AddColorStop(0, color.RGBA{15, 23, 42, 255})    // Slate 900
	grad.AddColorStop(0.5, color.RGBA{30, 27, 75, 255})  // Indigo 950
	grad.AddColorStop(1, color.RGBA{88, 28, 135, 255})   // Purple 900
	dc.SetFillStyle(grad)
	dc.DrawRectangle(0, 0, w, h)
	dc.Fill()
}

func (g *Generator) fetchImage(urlStr string) (image.Image, error) {
	fmt.Printf("[OG] Fetching image: %s\n", urlStr)
	
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	img, _, err := imageutil.Decode(resp.Body)
	return img, err
}


func (g *Generator) truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen-3]) + "..."
}

func (g *Generator) GenerateRankingOG(rankingType string, songs []domain.Song) (image.Image, error) {
	const (
		W = 1200
		H = 630
	)

	dc := gg.NewContext(W, H)
	
	// 1. Blurred background of the #1 song if available, otherwise fallback to gradient
	var bgUrl string
	if len(songs) > 0 && songs[0].Anime != nil {
		if songs[0].Anime.BannerUrl != nil {
			bgUrl = *songs[0].Anime.BannerUrl
		} else if songs[0].Anime.CoverUrl != nil {
			bgUrl = *songs[0].Anime.CoverUrl
		}
	}

	if bgUrl != "" {
		if err := g.drawBlurredBackground(dc, bgUrl, W, H); err != nil {
			drawGradientBackground(dc, W, H)
		}
	} else {
		drawGradientBackground(dc, W, H)
	}

	// 2. Artistic accents
	dc.SetRGBA(127.0/255.0, 19.0/255.0, 236.0/255.0, 0.15)
	dc.DrawCircle(0, 0, 500)
	dc.Fill()
	dc.DrawCircle(W, H, 400)
	dc.Fill()

	// 3. Branding
	if face, err := g.loadFont(g.fontBlack, 42); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.9)
		dc.DrawStringAnchored("ANIRANK", 80, 80, 0, 0.5)
	}

	// 4. Ranking Title & Subtitle
	badgeText := "GLOBAL RANKING"
	titleText := "Top Rated Songs"
	if rankingType == "seasonal" {
		badgeText = "SEASONAL RANKING"
		titleText = "Top Seasonal Hits"
	}

	// Badge
	if face, err := g.loadFont(g.fontBlack, 28); err == nil {
		dc.SetFontFace(face)
		dc.SetHexColor("#ff4e50")
		dc.DrawStringAnchored(badgeText, 80, 160, 0, 0.5)
	}

	// Title
	if face, err := g.loadFont(g.fontBlack, 68); err == nil {
		dc.SetFontFace(face)
		dc.SetRGB(1, 1, 1)
		dc.DrawStringWrapped(titleText, 80, 220, 0, 0, 450, 1.1, gg.AlignLeft)
	}

	// Subtitle/Description
	description := "The community's ultimate ranking of anime opening and ending themes."
	if face, err := g.loadFont(g.fontBold, 22); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.5)
		dc.DrawStringWrapped(description, 80, 380, 0, 0, 420, 1.3, gg.AlignLeft)
	}

	// Branding URL at the bottom left
	if face, err := g.loadFont(g.fontBold, 22); err == nil {
		dc.SetFontFace(face)
		dc.SetRGBA(1, 1, 1, 0.4)
		dc.DrawStringAnchored("ANIRANK.WORK", 80, H-80, 0, 0.5)
	}

	// 5. Draw Top 3 Songs on the right side
	listX := 560.0
	itemW := 560.0
	startY := 80.0
	spacingY := 160.0

	for i, song := range songs {
		if i >= 3 {
			break
		}
		y := startY + float64(i)*spacingY

		// Draw card background
		dc.SetRGBA(15.0/255.0, 23.0/255.0, 42.0/255.0, 0.6) // Semi-transparent card
		dc.DrawRoundedRectangle(listX, y, itemW, 140, 12)
		dc.Fill()

		// Draw card border
		dc.SetRGBA(255.0/255.0, 255.0/255.0, 255.0/255.0, 0.1)
		dc.SetLineWidth(1.5)
		dc.DrawRoundedRectangle(listX, y, itemW, 140, 12)
		dc.Stroke()

		// 5a. Rank Number
		rankColor := "#FFD700" // Gold
		if i == 1 {
			rankColor = "#C0C0C0" // Silver
		} else if i == 2 {
			rankColor = "#CD7F32" // Bronze
		}
		if face, err := g.loadFont(g.fontBlack, 48); err == nil {
			dc.SetFontFace(face)
			dc.SetHexColor(rankColor)
			dc.DrawStringAnchored(fmt.Sprintf("#%d", i+1), listX+40, y+70, 0.5, 0.5)
		}

		// 5b. Song Name
		songName := song.Name
		if songName == "" {
			songName = "N/A"
		}
		songName = g.truncate(songName, 32)
		
		if face, err := g.loadFont(g.fontBlack, 28); err == nil {
			dc.SetFontFace(face)
			dc.SetRGB(1, 1, 1)
			dc.DrawStringAnchored(songName, listX+100, y+45, 0, 0.5)
		}

		// 5c. Artists & Anime
		metaParts := []string{}
		if len(song.Artists) > 0 {
			metaParts = append(metaParts, g.truncate(song.Artists[0].Name, 20))
		}
		if song.Anime != nil {
			metaParts = append(metaParts, g.truncate(song.Anime.Title, 22))
		}
		metaText := strings.Join(metaParts, " • ")
		if face, err := g.loadFont(g.fontBold, 18); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.6)
			dc.DrawStringAnchored(metaText, listX+100, y+85, 0, 0.5)
		}

		// 5d. Score Badge (e.g. 9.4)
		if song.AverageRating > 0 {
			if face, err := g.loadFont(g.fontBlack, 26); err == nil {
				dc.SetFontFace(face)
				dc.SetHexColor("#FFD700")
				dc.DrawStringAnchored(fmt.Sprintf("%.1f", song.AverageRating), listX+itemW-45, y+55, 0.5, 0.5)
			}
			if face, err := g.loadFont(g.fontBold, 14); err == nil {
				dc.SetFontFace(face)
				dc.SetRGBA(1, 1, 1, 0.4)
				dc.DrawStringAnchored("SCORE", listX+itemW-45, y+95, 0.5, 0.5)
			}
		}
	}

	// If fewer than 3 songs (e.g., empty DB), draw a placeholder message on the right
	if len(songs) == 0 {
		dc.SetRGBA(15.0/255.0, 23.0/255.0, 42.0/255.0, 0.4)
		dc.DrawRoundedRectangle(listX, startY, itemW, H-startY-80, 16)
		dc.Fill()

		if face, err := g.loadFont(g.fontBold, 26); err == nil {
			dc.SetFontFace(face)
			dc.SetRGBA(1, 1, 1, 0.6)
			dc.DrawStringAnchored("No rankings calculated yet.", listX+itemW/2, startY+(H-startY-80)/2, 0.5, 0.5)
		}
	}

	return dc.Image(), nil
}
