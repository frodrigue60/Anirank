package admin

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"testing"

	"anirank/api/internal/domain"

	"github.com/gofiber/fiber/v2"
)

func TestApplyVideoMetadataFromForm(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("source", "DVD")
	_ = writer.WriteField("resolution", "480")
	_ = writer.WriteField("is_nc", "true")
	_ = writer.WriteField("is_bd", "false")
	_ = writer.WriteField("overlap", "None")
	_ = writer.Close()

	app := fiber.New()
	var captured domain.SongVariantVideo
	app.Post("/", func(c *fiber.Ctx) error {
		captured = domain.SongVariantVideo{}
		ApplyVideoMetadataFromForm(c, &captured)
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_, _ = app.Test(req)

	if captured.Source != "DVD" {
		t.Fatalf("expected source DVD, got %q", captured.Source)
	}
	if captured.Resolution != 480 {
		t.Fatalf("expected resolution 480, got %d", captured.Resolution)
	}
	if !captured.IsNC {
		t.Fatal("expected is_nc true")
	}
	if captured.Overlap != "None" {
		t.Fatalf("expected overlap None, got %q", captured.Overlap)
	}
}

func TestApplyVideoMetadataFromFormBDSourceSetsIsBD(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("source", "BD")
	_ = writer.WriteField("resolution", "1080")
	_ = writer.Close()

	app := fiber.New()
	var captured domain.SongVariantVideo
	app.Post("/", func(c *fiber.Ctx) error {
		captured = domain.SongVariantVideo{}
		ApplyVideoMetadataFromForm(c, &captured)
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_, _ = app.Test(req)

	if captured.Source != "BD" || !captured.IsBD {
		t.Fatalf("expected BD source with is_bd, got source=%q is_bd=%v", captured.Source, captured.IsBD)
	}
}

func TestMetadataTargetFromFormDefaultsToNew(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	app := fiber.New()
	var target string
	app.Post("/", func(c *fiber.Ctx) error {
		target = MetadataTargetFromForm(c)
		return c.SendStatus(204)
	})

	req := httptest.NewRequest("POST", "/", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	_, _ = app.Test(req)

	if target != "new" {
		t.Fatalf("expected new, got %q", target)
	}
}
