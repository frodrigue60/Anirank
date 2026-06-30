package admin

import (
	"testing"
)

func TestParseVideoTags(t *testing.T) {
	isNC, isBD, resolution, isUncensored, isSubbed, isLyrics, source, overlap := parseVideoTags("1080,NC")

	if !isNC {
		t.Error("expected isNC=true")
	}
	if isBD {
		t.Error("expected isBD=false for NC-only tags")
	}
	if resolution != 1080 {
		t.Errorf("expected resolution 1080, got %d", resolution)
	}
	if isUncensored || isSubbed || isLyrics {
		t.Error("unexpected optional flags")
	}
	if source != "TV" {
		t.Errorf("expected source TV, got %s", source)
	}
	if overlap != "None" {
		t.Errorf("expected overlap None, got %s", overlap)
	}

	_, isBD, resolution, _, _, isLyrics, source, overlap = parseVideoTags("720,BD,LYRICS,OVERLAP")
	if !isBD {
		t.Error("expected isBD=true")
	}
	if resolution != 720 {
		t.Errorf("expected resolution 720, got %d", resolution)
	}
	if !isLyrics {
		t.Error("expected isLyrics=true")
	}
	if source != "BD" {
		t.Errorf("expected source BD, got %s", source)
	}
	if overlap != "Overlap" {
		t.Errorf("expected overlap Overlap, got %s", overlap)
	}
}

func TestBuildVideosFromATInputs(t *testing.T) {
	videos := buildVideosFromATInputs([]ATVideoInput{
		{Path: "2025/winter/Foo-OP1.webm", Tags: "1080"},
		{Path: "", Tags: "720"},
		{Path: "videos/2025/winter/Bar-ED1.webm", Tags: "NC"},
	})

	if len(videos) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(videos))
	}
	if videos[0].VideoSrc == nil || *videos[0].VideoSrc != "videos/2025/winter/Foo-OP1.webm" {
		t.Errorf("unexpected first video_src: %v", videos[0].VideoSrc)
	}
	if videos[0].Resolution != 1080 {
		t.Errorf("expected resolution 1080, got %d", videos[0].Resolution)
	}
	if videos[1].VideoSrc == nil || *videos[1].VideoSrc != "videos/2025/winter/Bar-ED1.webm" {
		t.Errorf("unexpected second video_src: %v", videos[1].VideoSrc)
	}
	if !videos[1].IsNC {
		t.Error("expected isNC on second video")
	}
}
