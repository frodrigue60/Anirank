package admin

import (
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/pkg/rbac"
	"context"
	"github.com/gofiber/fiber/v2"
	"io"
)

type AdminUsecase struct {
	userAdmin    *UserAdminUsecase
	contentAdmin *ContentAdminUsecase
	adminRepo    domain.AdminRepository
	moderationR  domain.ModerationRepository
	jobsRepo     domain.JobsRepository
}

func NewAdminUsecase(
	ua *UserAdminUsecase,
	ca *ContentAdminUsecase,
	adminR domain.AdminRepository,
	modR domain.ModerationRepository,
	jr domain.JobsRepository,
) *AdminUsecase {
	return &AdminUsecase{
		userAdmin:    ua,
		contentAdmin: ca,
		adminRepo:    adminR,
		moderationR:  modR,
		jobsRepo:     jr,
	}
}

func (u *AdminUsecase) SnapshotRanking(ctx context.Context) error {
	return u.jobsRepo.SnapshotRankingPositions(ctx)
}

func (u *AdminUsecase) GetDashboardData(ctx context.Context) (*domain.DashboardStats, []domain.DailyMetric, error) {
	stats, err := u.adminRepo.GetDashboardStats(ctx)
	if err != nil {
		return nil, nil, err
	}

	metrics, err := u.adminRepo.GetDailyMetrics(ctx, 30) // Last 30 days
	if err != nil {
		return nil, nil, err
	}

	// Fetch Recent Activity
	pendingStatus := false
	recentReports, _ := u.moderationR.GetSongReports(ctx, &pendingStatus, 5, 0)
	recentRequests, _ := u.moderationR.GetUserRequests(ctx, &pendingStatus, 5, 0)

	stats.RecentReports = recentReports
	stats.RecentRequests = recentRequests

	return stats, metrics, nil
}

// ---- USERS ----
func (u *AdminUsecase) GetUsers(ctx context.Context, page, limit int, search string) ([]domain.User, int, error) {
	return u.userAdmin.GetUsers(ctx, page, limit, search)
}

func (u *AdminUsecase) GetRoles(ctx context.Context) ([]domain.Role, error) {
	return u.userAdmin.GetRoles(ctx)
}

func (u *AdminUsecase) GetUser(ctx context.Context, id uint64) (*domain.User, error) {
	return u.userAdmin.GetUser(ctx, id)
}

func (u *AdminUsecase) CreateUser(ctx context.Context, user *domain.User, roleIDs []uint64, badgeIDs []uint64, meta domain.AuditMetadata) error {
	return u.userAdmin.CreateUser(ctx, user, roleIDs, badgeIDs, meta)
}

func (u *AdminUsecase) UpdateUser(ctx context.Context, user *domain.User, roleIDs []uint64, badgeIDs []uint64, meta domain.AuditMetadata) error {
	return u.userAdmin.UpdateUser(ctx, user, roleIDs, badgeIDs, meta)
}

func (u *AdminUsecase) DeleteUser(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.userAdmin.DeleteUser(ctx, id, meta)
}

func (u *AdminUsecase) ResetPassword(ctx context.Context, id uint64) (string, error) {
	return u.userAdmin.ResetPassword(ctx, id)
}


// ---- ANIME ----
func (u *AdminUsecase) GetAnimes(ctx context.Context, page, limit int, search string, year, season, format, genre string, status *bool) ([]domain.Anime, int, error) {
	return u.contentAdmin.GetAnimes(ctx, page, limit, search, year, season, format, genre, status)
}

func (u *AdminUsecase) GetAnime(ctx context.Context, id uint64) (*domain.Anime, error) {
	return u.contentAdmin.GetAnime(ctx, id)
}

func (u *AdminUsecase) ToggleAnimeStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleAnimeStatus(ctx, id, meta)
}

func (u *AdminUsecase) CreateAnime(ctx context.Context, anime *domain.Anime, meta domain.AuditMetadata) error {
	return u.contentAdmin.CreateAnime(ctx, anime, meta)
}

func (u *AdminUsecase) UpdateAnime(ctx context.Context, anime *domain.Anime, meta domain.AuditMetadata) error {
	return u.contentAdmin.UpdateAnime(ctx, anime, meta)
}

func (u *AdminUsecase) DeleteAnime(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.DeleteAnime(ctx, id, meta)
}

func (u *AdminUsecase) BatchDeleteAnimes(ctx context.Context, ids []uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.BatchDeleteAnimes(ctx, ids, meta)
}

func (u *AdminUsecase) HandleAnimeImages(c *fiber.Ctx, anime *domain.Anime) {
	u.contentAdmin.HandleAnimeImages(c, anime)
}

func (u *AdminUsecase) BatchFetchAnimes(ctx context.Context, season string, year int, format string, meta domain.AuditMetadata) error {
	return u.contentAdmin.BatchFetchAnimes(ctx, season, year, format, meta)
}

func (u *AdminUsecase) SearchAnilistAnimes(ctx context.Context, query string, format string) ([]anilist.Media, error) {
	return u.contentAdmin.SearchAnilistAnimes(ctx, query, format)
}

func (u *AdminUsecase) ResolveAnimesURLs(animes []domain.Anime) {
	u.contentAdmin.ResolveAnimesURLs(animes)
}

func (u *AdminUsecase) ResolveAnimeURLs(anime *domain.Anime) {
	u.contentAdmin.ResolveAnimeURLs(anime)
}

func (u *AdminUsecase) ResolveArtistURLs(artist *domain.Artist) {
	u.contentAdmin.ResolveArtistURLs(artist)
}

func (u *AdminUsecase) ResolveArtistsURLs(artists []domain.Artist) {
	u.contentAdmin.ResolveArtistsURLs(artists)
}

func (u *AdminUsecase) HydrateAnimeSeason(ctx context.Context, year int, season string, meta domain.AuditMetadata, progress chan<- string) error {
	return u.contentAdmin.HydrateSeason(ctx, year, season, meta, progress)
}

func (u *AdminUsecase) SearchAnimeThemes(ctx context.Context, query string) ([]ATAnimeData, error) {
	return u.contentAdmin.SearchAnimeThemes(ctx, query)
}

func (u *AdminUsecase) HydrateAnimeThemes(ctx context.Context, ids []uint64, meta domain.AuditMetadata, progress chan<- string) error {
	return u.contentAdmin.HydrateAnimeThemes(ctx, ids, meta, progress)
}

func (u *AdminUsecase) ResolveSongsURLs(songs []domain.Song) {
	u.contentAdmin.ResolveSongsURLs(songs)
}

func (u *AdminUsecase) ResolveSongURLs(song *domain.Song) {
	u.contentAdmin.ResolveSongURLs(song)
}

func (u *AdminUsecase) CreateAnimeFromAnilist(ctx context.Context, anilistID int, meta domain.AuditMetadata) (*domain.Anime, error) {
	return u.contentAdmin.CreateAnimeFromAnilist(ctx, anilistID, meta)
}

func (u *AdminUsecase) SyncAnime(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.SyncAnime(ctx, id, meta)
}

func (u *AdminUsecase) SyncAnimeWithAnilist(ctx context.Context, anime *domain.Anime, media *anilist.Media, meta domain.AuditMetadata) error {
	pm := rbac.GetPermissionManager(u.contentAdmin.userRepo)
	return u.contentAdmin.SyncAnimeWithAnilist(ctx, anime, media, pm, meta)
}

func (u *AdminUsecase) BatchCreateAnimesFromAnilist(ctx context.Context, anilistIDs []int, meta domain.AuditMetadata) *domain.AnilistBatchImportResult {
	return u.contentAdmin.BatchCreateAnimesFromAnilist(ctx, anilistIDs, meta)
}

// ---- SONG ----
func (u *AdminUsecase) GetSongs(ctx context.Context, page, limit int, search string, animeID *uint64, status *bool) ([]domain.Song, int, error) {
	return u.contentAdmin.GetSongs(ctx, page, limit, search, animeID, status)
}

func (u *AdminUsecase) GetSong(ctx context.Context, id uint64) (*domain.Song, error) {
	return u.contentAdmin.GetSong(ctx, id)
}

func (u *AdminUsecase) CreateSong(ctx context.Context, s *domain.Song, meta domain.AuditMetadata) error {
	return u.contentAdmin.CreateSong(ctx, s, meta)
}

func (u *AdminUsecase) GetNextSongNumber(ctx context.Context, animeID uint64, songType string) (int, error) {
	return u.contentAdmin.GetNextSongNumber(ctx, animeID, songType)
}

func (u *AdminUsecase) UpdateSong(ctx context.Context, s *domain.Song, meta domain.AuditMetadata) error {
	return u.contentAdmin.UpdateSong(ctx, s, meta)
}

func (u *AdminUsecase) DeleteSong(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.DeleteSong(ctx, id, meta)
}

func (u *AdminUsecase) ToggleSongStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleSongStatus(ctx, id, meta)
}

func (u *AdminUsecase) SyncSongArtists(ctx context.Context, songID uint64, artistIDs []uint64) error {
	return u.contentAdmin.SyncSongArtists(ctx, songID, artistIDs)
}

func (u *AdminUsecase) SyncArtistsFromString(ctx context.Context, songID uint64, artistsStr string, meta domain.AuditMetadata) error {
	return u.contentAdmin.SyncArtistsFromString(ctx, songID, artistsStr, meta)
}

// ---- SONG VARIANT / VIDEO ----
func (u *AdminUsecase) GetVariants(ctx context.Context, page, limit int, search string, animeID *uint64, status *bool) ([]domain.SongVariant, int, error) {
	return u.contentAdmin.GetVariants(ctx, page, limit, search, animeID, status)
}

func (u *AdminUsecase) GetVariant(ctx context.Context, id uint64) (*domain.SongVariant, error) {
	return u.contentAdmin.GetVariant(ctx, id)
}

func (u *AdminUsecase) CreateVariant(ctx context.Context, v *domain.SongVariant, meta domain.AuditMetadata) error {
	return u.contentAdmin.CreateVariant(ctx, v, meta)
}

func (u *AdminUsecase) UpdateVariant(ctx context.Context, v *domain.SongVariant, meta domain.AuditMetadata) error {
	return u.contentAdmin.UpdateVariant(ctx, v, meta)
}

func (u *AdminUsecase) HandleVariantVideo(c *fiber.Ctx, v *domain.SongVariant) {
	u.contentAdmin.HandleVariantVideo(c, v)
}

func (u *AdminUsecase) DeleteVariant(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.DeleteVariant(ctx, id, meta)
}

func (u *AdminUsecase) ToggleVariantStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleVariantStatus(ctx, id, meta)
}


// ---- ARTISTS ----
func (u *AdminUsecase) GetArtists(ctx context.Context, page, limit int, search string) ([]domain.Artist, int, error) {
	return u.contentAdmin.GetArtists(ctx, page, limit, search)
}

func (u *AdminUsecase) GetArtist(ctx context.Context, id uint64) (*domain.Artist, error) {
	return u.contentAdmin.GetArtist(ctx, id)
}

func (u *AdminUsecase) CreateArtist(ctx context.Context, a *domain.Artist, meta domain.AuditMetadata) error {
	if err := u.contentAdmin.CreateArtist(ctx, a, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "created", a.ID, "artist", nil, a, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) GenerateArtistAvatar(ctx context.Context, artistID uint64, force bool) error {
	return u.contentAdmin.GenerateArtistAvatar(ctx, artistID, force)
}

func (u *AdminUsecase) BatchGenerateArtistAvatars(ctx context.Context, ids []uint64, progress chan<- string) error {
	return u.contentAdmin.BatchGenerateArtistAvatars(ctx, ids, progress)
}

func (u *AdminUsecase) SyncArtistAvatar(ctx context.Context, id uint64, meta domain.AuditMetadata) (*domain.Artist, error) {
	return u.contentAdmin.SyncArtistAvatar(ctx, id, meta)
}

func (u *AdminUsecase) RecountArtistStats(ctx context.Context, artistID *uint64, progress chan<- string) error {
	return u.contentAdmin.RecountArtistStats(ctx, artistID, progress)
}

func (u *AdminUsecase) UpdateArtist(ctx context.Context, a *domain.Artist, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.GetArtist(ctx, a.ID)
	if err := u.contentAdmin.UpdateArtist(ctx, a, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "updated", a.ID, "artist", existing, a, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) DeleteArtist(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.GetArtist(ctx, id)
	if err := u.contentAdmin.DeleteArtist(ctx, id, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "artist", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) ToggleArtistStatus(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleArtistStatus(ctx, id, meta)
}

func (u *AdminUsecase) MergeDuplicateArtists(ctx context.Context, progress chan<- string) error {
	return u.contentAdmin.MergeDuplicateArtists(ctx, progress)
}

func (u *AdminUsecase) UploadArtistAvatar(ctx context.Context, artistID uint64, file io.Reader, size int64, contentType string) error {
	return u.contentAdmin.UploadArtistAvatar(ctx, artistID, file, size, contentType)
}

// ---- TAXONOMIES ----
func (u *AdminUsecase) CreateYear(ctx context.Context, year *domain.Year, meta domain.AuditMetadata) error {
	if err := u.contentAdmin.CreateYear(ctx, year, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "created", year.ID, "taxonomy_year", nil, year, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) UpdateYear(ctx context.Context, year *domain.Year, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetYearByID(ctx, year.ID)
	if err := u.contentAdmin.UpdateYear(ctx, year, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "updated", year.ID, "taxonomy_year", existing, year, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) ToggleYearCurrent(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleYearCurrent(ctx, id, meta)
}

func (u *AdminUsecase) DeleteYear(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetYearByID(ctx, id)
	if err := u.contentAdmin.DeleteYear(ctx, id, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_year", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) GetYears(ctx context.Context) ([]domain.Year, error) {
	return u.contentAdmin.GetYears(ctx)
}

func (u *AdminUsecase) CreateSeason(ctx context.Context, season *domain.Season, meta domain.AuditMetadata) error {
	if err := u.contentAdmin.CreateSeason(ctx, season, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "created", season.ID, "taxonomy_season", nil, season, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) UpdateSeason(ctx context.Context, season *domain.Season, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetSeasonByID(ctx, season.ID)
	if err := u.contentAdmin.UpdateSeason(ctx, season, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "updated", season.ID, "taxonomy_season", existing, season, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) ToggleSeasonCurrent(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	return u.contentAdmin.ToggleSeasonCurrent(ctx, id, meta)
}

func (u *AdminUsecase) DeleteSeason(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetSeasonByID(ctx, id)
	if err := u.contentAdmin.DeleteSeason(ctx, id, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_season", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) GetSeasons(ctx context.Context) ([]domain.Season, error) {
	return u.contentAdmin.GetSeasons(ctx)
}

func (u *AdminUsecase) CreateFormat(ctx context.Context, format *domain.Format, meta domain.AuditMetadata) error {
	if err := u.contentAdmin.CreateFormat(ctx, format, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "created", format.ID, "taxonomy_format", nil, format, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) UpdateFormat(ctx context.Context, format *domain.Format, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetFormatByID(ctx, format.ID)
	if err := u.contentAdmin.UpdateFormat(ctx, format, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "updated", format.ID, "taxonomy_format", existing, format, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) DeleteFormat(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetFormatByID(ctx, id)
	if err := u.contentAdmin.DeleteFormat(ctx, id, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_format", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) GetFormats(ctx context.Context) ([]domain.Format, error) {
	return u.contentAdmin.GetFormats(ctx)
}

func (u *AdminUsecase) CreateGenre(ctx context.Context, genre *domain.Genre, meta domain.AuditMetadata) error {
	if err := u.contentAdmin.CreateGenre(ctx, genre, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "created", genre.ID, "taxonomy_genre", nil, genre, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) UpdateGenre(ctx context.Context, genre *domain.Genre, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetGenreByID(ctx, genre.ID)
	if err := u.contentAdmin.UpdateGenre(ctx, genre, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "updated", genre.ID, "taxonomy_genre", existing, genre, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) DeleteGenre(ctx context.Context, id uint64, meta domain.AuditMetadata) error {
	existing, _ := u.contentAdmin.taxonomyRepo.GetGenreByID(ctx, id)
	if err := u.contentAdmin.DeleteGenre(ctx, id, meta); err != nil {
		return err
	}
	_ = u.contentAdmin.auditUsecase.LogActions(ctx, meta.ActorID, "deleted", id, "taxonomy_genre", existing, nil, &meta.URL, &meta.IPAddress, &meta.UserAgent)
	return nil
}

func (u *AdminUsecase) SearchStudios(ctx context.Context, term string, limit int) ([]domain.Studio, error) {
	return u.contentAdmin.SearchStudios(ctx, term, limit)
}

func (u *AdminUsecase) SearchProducers(ctx context.Context, term string, limit int) ([]domain.Producer, error) {
	return u.contentAdmin.SearchProducers(ctx, term, limit)
}

func (u *AdminUsecase) SearchGenres(ctx context.Context, term string, limit int) ([]domain.Genre, error) {
	return u.contentAdmin.SearchGenres(ctx, term, limit)
}

func (u *AdminUsecase) GetAuditLogs(ctx context.Context, page, limit int, filters map[string]interface{}) ([]domain.AuditLog, int, error) {
	return u.contentAdmin.auditUsecase.GetAuditLogs(ctx, page, limit, filters)
}

func (u *AdminUsecase) GetAuditLog(ctx context.Context, id uint64) (*domain.AuditLog, error) {
	return u.contentAdmin.auditUsecase.GetAuditLog(ctx, id)
}

// ---- XP Activities ----
func (u *AdminUsecase) GetAllXPActivities(ctx context.Context) ([]domain.XPActivity, error) {
	return u.adminRepo.GetAllXPActivities(ctx)
}

func (u *AdminUsecase) UpdateXPActivity(ctx context.Context, activity *domain.XPActivity) error {
	return u.adminRepo.UpdateXPActivity(ctx, activity)
}

func (u *AdminUsecase) CheckAnilistStatus(ctx context.Context) (bool, string) {
	return u.contentAdmin.CheckAnilistStatus(ctx)
}

func (u *AdminUsecase) CheckAnimeThemesStatus(ctx context.Context) (bool, string) {
	return u.contentAdmin.CheckAnimeThemesStatus(ctx)
}
