package admin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/webhook"
)

type webhookUsecase struct {
	webhookRepo   domain.WebhookRepository
	animeRepo     domain.AnimeRepository
	songRepo      domain.SongRepository
	webhookClient webhook.Client
	mediaService  infrastructure.MediaService
}

func NewWebhookUsecase(
	wr domain.WebhookRepository,
	ar domain.AnimeRepository,
	sr domain.SongRepository,
	wc webhook.Client,
	ms infrastructure.MediaService,
) domain.WebhookUsecase {
	return &webhookUsecase{
		webhookRepo:   wr,
		animeRepo:     ar,
		songRepo:      sr,
		webhookClient: wc,
		mediaService:  ms,
	}
}

func (u *webhookUsecase) GetByUUID(ctx context.Context, uuid string) (*domain.Webhook, error) {
	return u.webhookRepo.GetByUUID(ctx, uuid)
}

func (u *webhookUsecase) GetAll(ctx context.Context) ([]domain.Webhook, error) {
	return u.webhookRepo.GetAll(ctx)
}

func (u *webhookUsecase) GetForContent(ctx context.Context, contentType string) ([]domain.Webhook, error) {
	return u.webhookRepo.GetByContentType(ctx, contentType)
}

func (u *webhookUsecase) Create(ctx context.Context, w *domain.Webhook) error {
	return u.webhookRepo.Create(ctx, w)
}

func (u *webhookUsecase) Update(ctx context.Context, w *domain.Webhook) error {
	return u.webhookRepo.Update(ctx, w)
}

func (u *webhookUsecase) Delete(ctx context.Context, uuid string) error {
	w, err := u.webhookRepo.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	return u.webhookRepo.Delete(ctx, w.ID)
}

func (u *webhookUsecase) TestWebhook(ctx context.Context, uuid string) error {
	w, err := u.webhookRepo.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"content": "🔔 **AniRank Webhook Test**\nThis is a test message to verify the webhook system integration.",
	}

	return u.webhookClient.SendJSON(ctx, w.URL, payload)
}

func (u *webhookUsecase) TriggerForAnime(ctx context.Context, webhookUUID string, animeID uint64) error {
	w, err := u.webhookRepo.GetByUUID(ctx, webhookUUID)
	if err != nil {
		return err
	}
	return u.triggerForAnime(ctx, w, animeID)
}

func (u *webhookUsecase) triggerForAnime(ctx context.Context, w *domain.Webhook, animeID uint64) error {
	anime, err := u.animeRepo.GetByID(ctx, animeID)
	if err != nil {
		return err
	}

	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = os.Getenv("FRONTEND_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	animeURL := fmt.Sprintf("%s/anime/%s", baseURL, anime.Slug)
	
	description := "Check out the latest themes available on AniRank!"
	if anime.Description != nil && *anime.Description != "" {
		description = *anime.Description
		if len(description) > 200 {
			description = description[:197] + "..."
		}
	}

	var coverURL string
	if anime.Cover != nil {
		coverURL = u.mediaService.GetURL(*anime.Cover)
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("📺 %s - Themes are now live!", anime.Title),
				"description": description,
				"url":         animeURL,
				"color":       15418782, // Orange/Peach
				"image": map[string]interface{}{
					"url": coverURL,
				},
				"footer": map[string]interface{}{
					"text": "AniRank Notifications",
				},
			},
		},
	}

	return u.webhookClient.SendJSON(ctx, w.URL, payload)
}

func (u *webhookUsecase) TriggerForSong(ctx context.Context, webhookUUID string, songID uint64) error {
	w, err := u.webhookRepo.GetByUUID(ctx, webhookUUID)
	if err != nil {
		return err
	}
	return u.triggerForSong(ctx, w, songID)
}

func (u *webhookUsecase) triggerForSong(ctx context.Context, w *domain.Webhook, songID uint64) error {
	song, err := u.songRepo.GetByID(ctx, songID)
	if err != nil {
		return err
	}

	anime, err := u.animeRepo.GetByID(ctx, song.AnimeID)
	if err != nil {
		return err
	}

	// Load artists for the song
	artists, _ := u.songRepo.GetArtistsBySongID(ctx, song.ID, false)
	artistNames := "Unknown Artist"
	if len(artists) > 0 {
		names := make([]string, len(artists))
		for i, a := range artists {
			names[i] = a.Name
		}
		artistNames = strings.Join(names, ", ")
	}

	// Resolve best song name
	songName := "Unknown Title"
	if song.SongRomaji != nil && *song.SongRomaji != "" {
		songName = *song.SongRomaji
	} else if song.SongEN != nil && *song.SongEN != "" {
		songName = *song.SongEN
	} else if song.SongJP != nil && *song.SongJP != "" {
		songName = *song.SongJP
	}

	// Resolve type name (Opening, Ending, etc)
	typeName := song.Type
	if typeName == "" {
		typeName = "Theme"
	}

	baseURL := os.Getenv("APP_URL")
	if baseURL == "" {
		baseURL = os.Getenv("FRONTEND_URL")
	}
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}

	songURL := fmt.Sprintf("%s/anime/%s/%s", baseURL, anime.Slug, song.Slug)

	var coverURL string
	if anime.Cover != nil {
		coverURL = u.mediaService.GetURL(*anime.Cover)
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       fmt.Sprintf("🎵 New %s: %s", typeName, songName),
				"description": fmt.Sprintf("Performed by **%s**\n\nNow available for **%s**!", artistNames, anime.Title),
				"url":         songURL,
				"color":       3447003, // Blue
				"thumbnail": map[string]interface{}{
					"url": coverURL,
				},
				"fields": []map[string]interface{}{
					{
						"name":   "Theme",
						"value":  fmt.Sprintf("%s %s", typeName, song.ThemeNum),
						"inline": true,
					},
				},
				"footer": map[string]interface{}{
					"text": "AniRank Notifications",
				},
			},
		},
	}

	return u.webhookClient.SendJSON(ctx, w.URL, payload)
}
func (u *webhookUsecase) NotifyNewAnime(ctx context.Context, animeID uint64) error {
	webhooks, err := u.webhookRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, wh := range webhooks {
		if !wh.IsActive {
			continue
		}
		supportsAnime := false
		for _, ct := range wh.ContentTypes {
			if ct == "animes" {
				supportsAnime = true
				break
			}
		}
		if supportsAnime {
			if err := u.triggerForAnime(ctx, &wh, animeID); err != nil {
				fmt.Printf("[Webhook] Error triggering for anime %d on webhook %s: %v\n", animeID, wh.Name, err)
			} else {
				fmt.Printf("[Webhook] Successfully notified anime %d to %s\n", animeID, wh.Name)
			}
		}
	}
	return nil
}

func (u *webhookUsecase) NotifyNewSong(ctx context.Context, songID uint64) error {
	webhooks, err := u.webhookRepo.GetAll(ctx)
	if err != nil {
		return err
	}

	for _, wh := range webhooks {
		if !wh.IsActive {
			continue
		}
		supportsSongs := false
		for _, ct := range wh.ContentTypes {
			if ct == "songs" {
				supportsSongs = true
				break
			}
		}
		if supportsSongs {
			if err := u.triggerForSong(ctx, &wh, songID); err != nil {
				fmt.Printf("[Webhook] Error triggering for song %d on webhook %s: %v\n", songID, wh.Name, err)
			} else {
				fmt.Printf("[Webhook] Successfully notified song %d to %s\n", songID, wh.Name)
			}
		}
	}
	return nil
}
 
func (u *webhookUsecase) NotifyCustomMessage(ctx context.Context, title, message string) error {
	webhooks, err := u.webhookRepo.GetAll(ctx)
	if err != nil {
		return err
	}
 
	for _, wh := range webhooks {
		if !wh.IsActive {
			continue
		}
		if err := u.sendCustomMessage(ctx, &wh, title, message); err != nil {
			fmt.Printf("[Webhook] Error sending custom message to %s: %v\n", wh.Name, err)
		}
	}
	return nil
}
 
func (u *webhookUsecase) SendCustomMessage(ctx context.Context, uuid string, title, message string) error {
	w, err := u.webhookRepo.GetByUUID(ctx, uuid)
	if err != nil {
		return err
	}
	return u.sendCustomMessage(ctx, w, title, message)
}
 
func (u *webhookUsecase) sendCustomMessage(ctx context.Context, w *domain.Webhook, title, message string) error {
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       "📢 " + title,
				"description": message,
				"color":       10181046, // Purple
				"footer": map[string]interface{}{
					"text": "AniRank Admin Notification",
				},
			},
		},
	}
 
	return u.webhookClient.SendJSON(ctx, w.URL, payload)
}
