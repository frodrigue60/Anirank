package v1

import (
	"fmt"
	"strings"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

type SEOHandler struct {
	seoUsecase domain.SEOUsecase
}

func NewSEOHandler(su domain.SEOUsecase) *SEOHandler {
	return &SEOHandler{
		seoUsecase: su,
	}
}

// GetMetadata renders a minimal HTML for bots.
func (h *SEOHandler) GetMetadata(c *fiber.Ctx) error {
	path := c.Params("*")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	seo, err := h.seoUsecase.GetMetadata(path)
	if err != nil {
		// Even on error, return something basic so bots don't fail
		return h.renderMinimal(c, "AniRank", "The Ultimate Anime Music Ranking Platform", "", "https://anirank.com"+path)
	}

	return h.renderMinimal(c, seo.Title, seo.Description, seo.Image, seo.URL)
}

func (h *SEOHandler) renderMinimal(c *fiber.Ctx, title, description, image, url string) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>%s</title>
    <meta name="description" content="%s">
    
    <!-- Open Graph -->
    <meta property="og:title" content="%s">
    <meta property="og:description" content="%s">
    <meta property="og:image" content="%s">
    <meta property="og:url" content="%s">
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="AniRank">
    <meta property="og:locale" content="en_US">
    <meta property="og:logo" content="https://anirank.work/favicon.png">
    
    <!-- Twitter -->
    <meta name="twitter:card" content="summary_large_image">
    <meta name="twitter:title" content="%s">
    <meta name="twitter:description" content="%s">
    <meta name="twitter:image" content="%s">

    <!-- SEO Meta -->
    <link rel="canonical" href="%s">
</head>
<body>
    <h1>%s</h1>
    <p>%s</p>
</body>
</html>`, 
		title, description, 
		title, description, image, url, 
		title, description, image, 
		url, title, description)

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
