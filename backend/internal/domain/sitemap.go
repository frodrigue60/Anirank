package domain

import "time"

type SitemapItem struct {
	Loc        string    `json:"loc"`
	LastMod    time.Time `json:"lastmod"`
	ChangeFreq string    `json:"changefreq"`
	Priority   float64   `json:"priority"`
}
