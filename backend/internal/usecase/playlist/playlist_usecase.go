package playlist

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"context"
	"regexp"
)

var iframeSrcRegex = regexp.MustCompile(`src="([^"]+)"`)

type PlaylistUsecase struct {
	playlistRepo    domain.PlaylistRepository
	songRepo        domain.SongRepository
	animeRepo       domain.AnimeRepository
	interactionRepo domain.InteractionRepository
	mediaService    infrastructure.MediaService
	xpUsecase       domain.XPUsecase
}

func NewPlaylistUsecase(pr domain.PlaylistRepository, sr domain.SongRepository, ar domain.AnimeRepository, ir domain.InteractionRepository, media infrastructure.MediaService, xu domain.XPUsecase) *PlaylistUsecase {
	return &PlaylistUsecase{
		playlistRepo:    pr,
		songRepo:        sr,
		animeRepo:       ar,
		interactionRepo: ir,
		mediaService:    media,
		xpUsecase:       xu,
	}
}

// GetUserPlaylists returns all playlists for a specific user.
// If the requesting user is the owner, include private playlists.
func (u *PlaylistUsecase) GetUserPlaylists(ctx context.Context, requestingUserID *uint64, targetUserID uint64, limit, offset int) ([]domain.Playlist, error) {
	includePrivate := false
	if requestingUserID != nil && *requestingUserID == targetUserID {
		includePrivate = true
	}
	playlists, err := u.playlistRepo.GetByUserID(ctx, targetUserID, includePrivate, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range playlists {
		u.enrichPlaylist(&playlists[i])
	}

	return playlists, nil
}

func (u *PlaylistUsecase) GetMyPlaylists(ctx context.Context, userID, songID uint64, limit, offset int) ([]domain.Playlist, error) {
	playlists, err := u.playlistRepo.GetByUserIDWithSongCheck(ctx, userID, songID, true, limit, offset)
	if err != nil {
		return nil, err
	}

	for i := range playlists {
		u.enrichPlaylist(&playlists[i])
	}

	return playlists, nil
}

func (u *PlaylistUsecase) GetPlaylist(ctx context.Context, playlistID uint64, requestingUserID *uint64) (*domain.Playlist, error) {
	playlist, err := u.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	// Privacy check
	if !playlist.IsPublic {
		if requestingUserID == nil || *requestingUserID != playlist.UserID {
			return nil, domain.NewAppError(403, "This playlist is private", nil)
		}
	}

	return playlist, nil
}

func (u *PlaylistUsecase) CreatePlaylist(ctx context.Context, userID uint64, name string, description *string, isPublic bool) (*domain.Playlist, error) {
	if len(name) < 3 || len(name) > 100 {
		return nil, domain.NewAppError(400, "Playlist name must be between 3 and 100 characters", nil)
	}

	playlist := &domain.Playlist{
		UserID:      userID,
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
	}

	err := u.playlistRepo.Create(ctx, playlist)
	if err != nil {
		return nil, domain.NewAppError(500, "Could not create playlist", err)
	}

	_ = u.xpUsecase.AwardXP(ctx, userID, "create_playlist", map[string]interface{}{"playlist_id": playlist.ID})
	return playlist, nil
}

func (u *PlaylistUsecase) UpdatePlaylist(ctx context.Context, userID, playlistID uint64, name string, description *string, isPublic bool) error {
	if len(name) < 3 || len(name) > 100 {
		return domain.NewAppError(400, "Playlist name must be between 3 and 100 characters", nil)
	}

	playlist := &domain.Playlist{
		ID:          playlistID,
		UserID:      userID,
		Name:        name,
		Description: description,
		IsPublic:    isPublic,
	}

	err := u.playlistRepo.Update(ctx, playlist)
	if err != nil {
		return domain.NewAppError(403, "You cannot update this playlist", err) // Maps 'no rows' to 403 generally when checking combined ID/User
	}
	return nil
}

func (u *PlaylistUsecase) DeletePlaylist(ctx context.Context, userID, playlistID uint64) error {
	err := u.playlistRepo.Delete(ctx, playlistID, userID)
	if err != nil {
		return domain.NewAppError(403, "You cannot delete this playlist", err)
	}
	return nil
}

// ---- Songs Management ----

func (u *PlaylistUsecase) AddSongToPlaylist(ctx context.Context, userID, playlistID, songID uint64, position int) error {
	// 1. Verify ownership of playlist
	playlist, err := u.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return domain.NewAppError(403, "Unauthorized", nil)
	}

	// 2. Add song
	err = u.playlistRepo.AddSong(ctx, playlistID, songID, position)
	if err == nil {
		_ = u.xpUsecase.AwardXP(ctx, userID, "add_to_playlist", map[string]interface{}{"song_id": songID, "playlist_id": playlistID})
	}
	return err
}

func (u *PlaylistUsecase) RemoveSongFromPlaylist(ctx context.Context, userID, playlistID, songID uint64) error {
	// 1. Verify ownership
	playlist, err := u.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return domain.NewAppError(403, "Unauthorized", nil)
	}

	// 2. Remove song
	return u.playlistRepo.RemoveSong(ctx, playlistID, songID)
}

func (u *PlaylistUsecase) ReorderPlaylistSongs(ctx context.Context, userID, playlistID uint64, positions map[uint64]int) error {
	// 1. Verify ownership
	playlist, err := u.playlistRepo.GetByID(ctx, playlistID)
	if err != nil {
		return err
	}
	if playlist.UserID != userID {
		return domain.NewAppError(403, "Unauthorized", nil)
	}

	// 2. Build bulk update
	var items []domain.PlaylistSong
	for songID, pos := range positions {
		items = append(items, domain.PlaylistSong{
			PlaylistID: playlistID,
			SongID:     songID,
			Position:   pos,
		})
	}

	return u.playlistRepo.UpdateSongPositions(ctx, playlistID, items)
}

func (u *PlaylistUsecase) GetPlaylistSongs(ctx context.Context, playlistID uint64, requestingUserID *uint64) ([]domain.PlaylistSong, error) {
	// Verify view permission first
	_, err := u.GetPlaylist(ctx, playlistID, requestingUserID)
	if err != nil {
		return nil, err
	}

	items, err := u.playlistRepo.GetSongs(ctx, playlistID)
	if err != nil {
		return nil, err
	}

	// Enrich each song
	for i := range items {
		if items[i].Song != nil {
			u.enrichPlaylistSong(ctx, items[i].Song, requestingUserID)
		}
	}

	return items, nil
}

func (u *PlaylistUsecase) enrichPlaylistSong(ctx context.Context, s *domain.Song, userID *uint64) {
	// Load anime with images
	if s.Anime == nil && s.AnimeID > 0 {
		anime, _ := u.animeRepo.GetByID(ctx, s.AnimeID)
		if anime != nil {
			u.animeRepo.LoadRelations(ctx, anime, false)
		}
		s.Anime = anime
	}

	// Load Artists
	if len(s.Artists) == 0 {
		artists, _ := u.songRepo.GetArtistsBySongID(ctx, s.ID)
		for i := range artists {
			if artists[i].Avatar != nil {
				artists[i].AvatarUrl = u.mediaService.Resolve(artists[i].Avatar)
			}
		}
		s.Artists = artists
	}

	// Load Variants with videos
	if len(s.Variants) == 0 {
		variants, _ := u.songRepo.GetVariantsBySongID(ctx, s.ID)
		for vi := range variants {
			if variants[vi].Video != nil && variants[vi].Video.EmbedUrl != nil {
				matches := iframeSrcRegex.FindStringSubmatch(*variants[vi].Video.EmbedUrl)
				if len(matches) > 1 {
					variants[vi].Video.EmbedUrl = &matches[1]
				}
			}
		}
		s.Variants = variants
	}

	// Interaction Flags
	if userID != nil && u.interactionRepo != nil {
		// Favorite
		fav, _ := u.interactionRepo.IsFavoritedByUser(ctx, *userID, s.ID, domain.TypeSong)
		s.IsFavorited = fav

		// Reaction (Like/Dislike)
		react, err := u.interactionRepo.GetReactionByUser(ctx, *userID, s.ID, domain.TypeSong)
		if err == nil && react != nil {
			s.IsLiked = react.Type == 1
			s.IsDisliked = react.Type == -1
		}
	}

	// Compute name
	if s.SongRomaji != nil {
		s.Name = *s.SongRomaji
	} else if s.SongEN != nil {
		s.Name = *s.SongEN
	} else if s.SongJP != nil {
		s.Name = *s.SongJP
	}

	// Set image URLs
	if s.Anime != nil {
		s.Anime.CoverUrl = u.mediaService.Resolve(s.Anime.Cover)
		s.Anime.BannerUrl = u.mediaService.Resolve(s.Anime.Banner)
	}

	// Fetch Average Rating
	if u.interactionRepo != nil {
		avg, _ := u.interactionRepo.GetAverageRating(ctx, s.ID)
		s.AverageRating = avg
	}
}

func (u *PlaylistUsecase) enrichPlaylist(p *domain.Playlist) {
	p.BannerUrl = u.mediaService.Resolve(p.LatestBanner)
}
