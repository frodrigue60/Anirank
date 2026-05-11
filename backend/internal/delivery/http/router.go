package http

import (
	"anirank/api/internal/delivery/http/middleware"
	"anirank/api/internal/domain"
	v1 "anirank/api/internal/delivery/http/v1"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/repository/postgres"
	"anirank/api/internal/usecase/admin"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/interaction"
	"anirank/api/internal/usecase/moderation"
	"anirank/api/internal/usecase/playlist"
	"anirank/api/internal/usecase/public"
	"anirank/api/internal/infrastructure/og"
	"anirank/api/internal/usecase/tournament"
	"anirank/api/internal/usecase/announcement"
	"anirank/api/internal/usecase/notification"
	"time"
	_ "anirank/api/docs" // Import swagger docs

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
	"github.com/jmoiron/sqlx"
)

func SetupPublicRoutes(app *fiber.App,
	db *sqlx.DB,
	discoveryUsecase *public.DiscoveryUsecase,
	animeUsecase *public.AnimeUsecase,
	searchUsecase *public.SearchUsecase,
	catalogUsecase *public.CatalogUsecase,
	authUsecase *auth.AuthUsecase,
	interactionUsecase *interaction.InteractionUsecase,
	playlistUsecase *playlist.PlaylistUsecase,
	adminUsecase *admin.AdminUsecase,
	moderationUsecase *moderation.ModerationUsecase,
	tournamentUsecase *tournament.TournamentUsecase,
	auditLogUsecase domain.AuditLogUsecase,
	jwtService *auth.JWTService,
	storageService infrastructure.StorageService,
	mediaService infrastructure.MediaService,
	xpUsecase domain.XPUsecase,
	activityUsecase domain.ActivityUsecase,
	statsUsecase domain.StatsUsecase,
	ogGenerator *og.Generator,
	shareHandler *v1.ShareHandler,
	seoHandler *v1.SEOHandler,
	storage fiber.Storage,
	appCache domain.Cache) {

	// Core Repositories
	songRepo := postgres.NewSongRepository(db)
	userRepo := postgres.NewUserRepository(db)
	animeRepo := postgres.NewAnimeRepository(db)
	artistRepo := postgres.NewArtistRepository(db)
	playlistRepo := postgres.NewPlaylistRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	interactionRepo := postgres.NewInteractionRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	notificationUsecase := notification.NewNotificationUsecase(notificationRepo, appCache)

	// HTTP Handlers
	discoveryHandler := v1.NewDiscoveryHandler(discoveryUsecase)
	animeHandler := v1.NewAnimeHandler(animeUsecase)
	searchHandler := v1.NewSearchHandler(searchUsecase)
	catalogHandler := v1.NewCatalogHandler(catalogUsecase)
	authHandler := v1.NewAuthHandler(authUsecase)
	
	badgeRepo := postgres.NewBadgeRepository(db)
	badgeUsecase := admin.NewBadgeUsecase(badgeRepo, userRepo, interactionRepo, commentRepo, storageService, auditLogUsecase)

	interactionUsecase = interaction.NewInteractionUsecase(interactionRepo, commentRepo, userRepo, notificationUsecase, songRepo, animeRepo, artistRepo, mediaService, xpUsecase, activityUsecase, badgeUsecase, moderationUsecase)
	interactionHandler := v1.NewInteractionHandler(interactionUsecase, activityUsecase, songRepo, userRepo, animeRepo, artistRepo, commentRepo)
	playlistHandler := v1.NewPlaylistHandler(playlistUsecase, playlistRepo, songRepo, userRepo)
	adminHandler := v1.NewAdminHandler(adminUsecase, songRepo, userRepo, animeRepo, artistRepo, playlistRepo)
	moderationHandler := v1.NewModerationHandler(moderationUsecase, songRepo, commentRepo, userRepo)
	tournamentHandler := v1.NewTournamentHandler(tournamentUsecase)

	announcementRepo := postgres.NewAnnouncementRepository(db)
	announcementUsecase := announcement.NewAnnouncementUsecase(announcementRepo, mediaService)
	announcementHandler := v1.NewAnnouncementHandler(announcementUsecase)

	badgeHandler := v1.NewBadgeHandler(badgeUsecase)

	activityHandler := v1.NewActivityHandler(activityUsecase)

	notificationHandler := v1.NewNotificationHandler(notificationUsecase)

	statsHandler := v1.NewStatsHandler(statsUsecase)
	ogHandler := v1.NewOGHandler(ogGenerator, animeUsecase, catalogUsecase, playlistUsecase, statsUsecase)

	permissionUsecase := admin.NewPermissionUsecase(userRepo)
	permissionHandler := v1.NewPermissionHandler(permissionUsecase)

	// API V1 Group
	api := app.Group("/api")

	// --- SWAGGER DOCS ---
	api.Get("/swagger/*", swagger.HandlerDefault)

	// --- HEALTH CHECK ---
	api.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// --- PUBLIC ROUTES ---
	authLimiter := middleware.NewAuthLimiter(storage)
	publicLimiter := middleware.NewPublicApiLimiter(storage)
	csrfMiddleware := middleware.NewCSRFMiddleware(storage)

	// Apply CSRF to all API routes
	api.Use(csrfMiddleware)

	// Auth
	api.Post("/login", authLimiter, authHandler.Login)
	api.Post("/register", authLimiter, authHandler.Register)
	api.Post("/forgot-password", authLimiter, authHandler.ForgotPassword)
	api.Post("/reset-password", authLimiter, authHandler.ResetPassword)
	api.Get("/verify-email", authLimiter, authHandler.VerifyEmail)
	api.Get("/auth/google/login", authHandler.GoogleLogin)
	api.Post("/auth/google/login-callback", authLimiter, authHandler.GoogleLoginCallback)
	api.Get("/auth/anilist/login", authHandler.AnilistLogin)
	api.Post("/auth/anilist/login-callback", authLimiter, authHandler.AnilistLoginCallback)
	api.Post("/auth/anilist/register", authLimiter, authHandler.AnilistRegister)
	api.Get("/auth/discord/login", authHandler.DiscordLogin)
	api.Post("/auth/discord/login-callback", authLimiter, authHandler.DiscordLoginCallback)

	// Init data for SPA
	api.Get("/init", middleware.NewResponseCache(storage, 5*time.Minute), discoveryHandler.Init)

	// Site Statistics
	api.Get("/site-statistics", middleware.NewResponseCache(storage, 5*time.Minute), statsHandler.GetSiteStats)

	catalogApi := api.Group("", publicLimiter)

	// Global Search Engine
	catalogApi.Get("/search", searchHandler.Search)

	// Anime Endpoints
	catalogApi.Get("/animes", middleware.NewResponseCache(storage, 5*time.Minute), animeHandler.Index)
	catalogApi.Get("/animes/:slug", middleware.NewResponseCache(storage, 5*time.Minute), animeHandler.Show)

	// Catalog: Songs
	catalogApi.Get("/songs", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.SongIndex)
	catalogApi.Get("/songs/ranking/:type", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.SongRanking)
	catalogApi.Get("/songs/:uuid/comments", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), interactionHandler.GetSongComments)
	catalogApi.Get("/songs/:anime_slug/:song_slug", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.SongShow)
	catalogApi.Get("/animes/:anime_slug/songs/:song_slug", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.SongShow)

	// Catalog: Artists
	catalogApi.Get("/artists", middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.ArtistIndex)
	catalogApi.Get("/artists/:slug", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.ArtistShow)
	catalogApi.Get("/artists/:slug/songs", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.ArtistShow)

	// Catalog: Studios
	catalogApi.Get("/studios", catalogHandler.StudioIndex)
	catalogApi.Get("/studios/:slug", catalogHandler.StudioShow)

	// Catalog: Producers
	catalogApi.Get("/producers", catalogHandler.ProducerIndex)
	catalogApi.Get("/producers/:slug", catalogHandler.ProducerShow)

	// Catalog: Users
	catalogApi.Get("/users/ranking", catalogHandler.UserRanking)
	catalogApi.Get("/users/:slug", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), catalogHandler.UserProfile)
	// Catalog: Home
	catalogApi.Get("/home", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), middleware.NewResponseCache(storage, 5*time.Minute), catalogHandler.Home)

	// Catalog: Playlists
	catalogApi.Get("/playlists", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), catalogHandler.PlaylistIndex)

	// Sitemap
	catalogApi.Get("/catalog/sitemap", catalogHandler.GetSitemap)
	catalogApi.Get("/catalog/sitemap.xml", catalogHandler.GetSitemapXML)

	// Activity Feed
	api.Get("/activities", activityHandler.Index)
	api.Get("/activities/recent", activityHandler.Recent)

	// Announcements Public
	api.Get("/announcements", announcementHandler.GetPublicAnnouncements)

	// Tournaments Public
	api.Get("/tournaments", tournamentHandler.ListTournamentsPublic)
	api.Get("/tournaments/active", tournamentHandler.GetActiveTournament)
	api.Get("/tournaments/:slug", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), tournamentHandler.GetTournamentBySlug)

	// Public Interactions (e.g Feed)
	api.Get("/interactions/feed", interactionHandler.Feed)

	// Shared Proxy / Share Routes (Bot Proxy)
	api.Get("/share/anime/:slug", shareHandler.AnimeShare)
	api.Get("/share/song/:anime_slug/:song_slug", shareHandler.SongShare)
	api.Get("/share/artist/:slug", shareHandler.ArtistShare)
	api.Get("/share/playlist/:id", shareHandler.PlaylistShare)
	api.Get("/share/user/:slug", shareHandler.UserShare)

	// OG Images
	api.Get("/og/anime/:slug", ogHandler.AnimeOG)
	api.Get("/og/song/:anime_slug/:song_slug", ogHandler.SongOG)
	api.Get("/og/artist/:slug", ogHandler.ArtistOG)
	api.Get("/og/playlist/:pid", ogHandler.PlaylistOG)
	api.Get("/og/user/:slug", ogHandler.UserOG)
	api.Get("/og/home", ogHandler.HomeOG)
	
	// SEO Bot Proxy route
	app.Get("/seo-bot/*", seoHandler.GetMetadata)


	// Comments Public
	api.Get("/comments", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), interactionHandler.GetComments)
	api.Get("/comments/:id/replies", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), interactionHandler.GetReplies)

	// User Public Playlists
	api.Get("/playlists/users/:id", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), playlistHandler.GetUserPlaylists)
	api.Get("/playlists/:id", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), playlistHandler.GetPlaylist)

	// User Public Data
	api.Get("/users/:slug/playlists", middleware.OptionalAuthMiddleware(jwtService, userRepo, appCache), catalogHandler.UserPlaylists)
	api.Get("/users/:slug/followers", catalogHandler.UserFollowers)
	api.Get("/users/:slug/following", catalogHandler.UserFollowing)
	api.Get("/users/:slug/anilist-list", catalogHandler.UserAnilistList)
	api.Post("/users/favorites/themes", catalogHandler.UserFavorites)
	api.Post("/users/favorites/artists", catalogHandler.UserArtistFavorites)

	// --- PROTECTED ROUTES ---
	protected := api.Group("/", middleware.AuthMiddleware(jwtService, userRepo, appCache))
	vRequired := middleware.VerifiedMiddleware()

	// User Profile
	protected.Get("/profile", authHandler.Profile)
	protected.Post("/users/avatar", vRequired, authHandler.UpdateAvatar)
	protected.Post("/users/banner", vRequired, authHandler.UpdateBanner)
	protected.Post("/users/score-format", authHandler.UpdateScoreFormat)
	protected.Put("/users/email", authHandler.UpdateEmail)
	protected.Post("/users/resend-verification", authHandler.ResendVerification)
	protected.Patch("/users/profile", vRequired, authHandler.UpdateProfile)

	// Anilist Link
	protected.Get("/auth/anilist/link", vRequired, authHandler.AnilistLink)
	protected.Post("/auth/anilist/callback", vRequired, authHandler.AnilistCallback)

	// Google Link
	protected.Get("/auth/google/link", vRequired, authHandler.GoogleLink)
	protected.Post("/auth/google/callback", vRequired, authHandler.GoogleCallback)

	// Discord Link
	protected.Delete("/auth/:provider/unlink", vRequired, authHandler.UnlinkSocial)
	protected.Get("/auth/discord/link", vRequired, authHandler.DiscordLink)
	protected.Post("/auth/discord/callback", vRequired, authHandler.DiscordCallback)

	protected.Post("/interactions/ratings", vRequired, interactionHandler.Rate)
	protected.Post("/interactions/reactions", vRequired, interactionHandler.React)
	protected.Post("/interactions/favorites", vRequired, interactionHandler.ToggleFavorite)
	protected.Post("/comments", vRequired, interactionHandler.SongComment)
	protected.Put("/comments/:id", vRequired, interactionHandler.UpdateComment)
	protected.Delete("/comments/:id", vRequired, interactionHandler.DeleteComment)

	protected.Post("/users/:id/follow", vRequired, interactionHandler.FollowUser)
	protected.Delete("/users/:id/follow", vRequired, interactionHandler.UnfollowUser)
	protected.Get("/users/:id/is-following", interactionHandler.IsFollowing)

	// Notifications
	protected.Get("/notifications/stream", notificationHandler.Stream)
	protected.Get("/notifications", notificationHandler.Index)
	protected.Put("/notifications/:id/read", notificationHandler.MarkAsRead)
	protected.Post("/notifications/read-all", notificationHandler.MarkAllAsRead)
	protected.Delete("/notifications/:id", notificationHandler.Delete)
	protected.Get("/notifications/unread-count", notificationHandler.GetUnreadCount)

	// Settings (Categorized under notifications as they are specific to them)
	protected.Get("/settings/notifications", notificationHandler.GetSettings)
	protected.Put("/settings/notifications", notificationHandler.UpdateSettings)

	// --- PROTECTED PLAYLISTS ---
	protected.Get("/me/playlists", playlistHandler.GetMyPlaylists)
	protected.Post("/playlists", vRequired, playlistHandler.Create)
	protected.Put("/playlists/:id", vRequired, playlistHandler.Update)
	protected.Delete("/playlists/:id", vRequired, playlistHandler.Delete)
	protected.Post("/playlists/:id/songs", vRequired, playlistHandler.AddSong)
	protected.Delete("/playlists/:id/songs/:songID", vRequired, playlistHandler.RemoveSong)
	protected.Put("/playlists/:id/songs/reorder", vRequired, playlistHandler.ReorderSongs)

	// Protected Moderation / User Support
	protected.Post("/songs/reports", vRequired, moderationHandler.CreateSongReport)
	protected.Post("/comments/reports", vRequired, moderationHandler.CreateCommentReport)
	protected.Post("/users/reports", vRequired, moderationHandler.CreateUserReport)
	protected.Post("/user-requests", vRequired, moderationHandler.CreateUserRequest)

	// Protected Tournament Voting
	protected.Post("/tournaments/matchups/:id/vote", vRequired, tournamentHandler.SubmitVote)

	// --- STAFF EXCLUSIVE ROUTES ---
	adminOnly := protected.Group("/admin", middleware.StaffMiddleware())

	// Dashboard & System
	adminOnly.Get("/dashboard", adminHandler.GetDashboard)
	adminOnly.Get("/system/api-status", adminHandler.GetApiStatus)
	adminOnly.Post("/og/flush", ogHandler.FlushOGCache)

	// Admin Moderation Tickets Review
	adminOnly.Get("/songs/reports", moderationHandler.GetSongReports)
	adminOnly.Get("/songs/reports/:id", moderationHandler.GetSongReport)
	adminOnly.Put("/songs/reports/:id/resolve", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.ResolveSongReport)
	adminOnly.Delete("/songs/reports/:id", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.DeleteSongReport)
	
	adminOnly.Get("/comments/reports", moderationHandler.GetCommentReports)
	adminOnly.Get("/comments/reports/:id", moderationHandler.GetCommentReport)
	adminOnly.Put("/comments/reports/:id/resolve", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.ResolveCommentReport)
	adminOnly.Delete("/comments/reports/:id", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.DeleteCommentReport)
	adminOnly.Get("/user-requests", moderationHandler.GetUserRequests)
	adminOnly.Get("/user-requests/:id", moderationHandler.GetUserRequest)
	adminOnly.Patch("/user-requests/:id/status", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.UpdateUserRequestStatus)
	adminOnly.Delete("/user-requests/:id", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.DeleteUserRequest)

	adminOnly.Get("/users/reports", moderationHandler.GetUserReports)
	adminOnly.Get("/users/reports/:id", moderationHandler.GetUserReport)
	adminOnly.Put("/users/reports/:id/resolve", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.ResolveUserReport)
	adminOnly.Delete("/users/reports/:id", middleware.HasPermissionMiddleware("reports.manage", userRepo), moderationHandler.DeleteUserReport)

	// Ranking Operations
	adminOnly.Post("/ranking/snapshot", adminHandler.SnapshotRankingPositions)

	// Audit Logs
	adminOnly.Get("/audit-logs", adminHandler.GetAuditLogs)
	adminOnly.Get("/audit-logs/:id", adminHandler.GetAuditLog)

	// XP Activities
	adminOnly.Get("/xp-activities", adminHandler.GetXPActivities)
	adminOnly.Put("/xp-activities/:id", adminHandler.UpdateXPActivity)

	// Users Operations
	adminOnly.Get("/users", adminHandler.GetUsers)
	adminOnly.Post("/users", middleware.HasPermissionMiddleware("users.manage", userRepo), adminHandler.CreateUser)
	adminOnly.Get("/roles", adminHandler.GetRoles)
	adminOnly.Get("/users/:id", adminHandler.GetUser)
	adminOnly.Put("/users/:id", middleware.HasPermissionMiddleware("users.manage", userRepo), adminHandler.UpdateUser)
	adminOnly.Delete("/users/:id", middleware.HasPermissionMiddleware("users.manage", userRepo), adminHandler.DeleteUser)
	adminOnly.Post("/users/:id/reset-password", middleware.HasPermissionMiddleware("users.manage", userRepo), adminHandler.ResetPassword)

	// Anime Operations
	adminOnly.Get("/animes", adminHandler.GetAnimes)
	adminOnly.Get("/animes/anilist-search", adminHandler.SearchAnilist)
	adminOnly.Post("/animes/from-anilist", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.CreateAnimeFromAnilist)
	adminOnly.Post("/animes/batch-from-anilist", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.BatchCreateAnimesFromAnilist)
	adminOnly.Get("/animes/:id", adminHandler.GetAnime)
	adminOnly.Post("/animes", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.CreateAnime)
	adminOnly.Post("/animes/batch", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.BatchFetchAnimes)
	adminOnly.Post("/animes/hydrate", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.HydrateAnimeSeason)
	adminOnly.Get("/animes/animethemes/search", adminHandler.SearchAnimeThemes)
	adminOnly.Post("/animes/animethemes/hydrate", middleware.HasPermissionMiddleware("anime.create", userRepo), adminHandler.HydrateAnimeThemes)
	adminOnly.Put("/animes/:id", middleware.HasPermissionMiddleware("anime.edit", userRepo), adminHandler.UpdateAnime)
	adminOnly.Patch("/animes/:id/status", middleware.HasPermissionMiddleware("anime.edit", userRepo), adminHandler.ToggleAnimeStatus)
	adminOnly.Post("/animes/:id/sync", middleware.HasPermissionMiddleware("anime.edit", userRepo), adminHandler.SyncAnime)
	adminOnly.Delete("/animes/:id", middleware.HasPermissionMiddleware("anime.delete", userRepo), adminHandler.DeleteAnime)
	adminOnly.Post("/animes/batch-delete", middleware.HasPermissionMiddleware("anime.delete", userRepo), adminHandler.BatchDeleteAnimes)

	// Songs Group
	adminOnly.Get("/songs", adminHandler.GetSongs)
	adminOnly.Get("/songs/latest-number", adminHandler.GetLatestSongNumber)
	adminOnly.Get("/songs/:id", adminHandler.GetSong)
	adminOnly.Post("/songs", middleware.HasPermissionMiddleware("song.create", userRepo), adminHandler.CreateSong)
	adminOnly.Put("/songs/:id", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.UpdateSong)
	adminOnly.Delete("/songs/:id", middleware.HasPermissionMiddleware("song.delete", userRepo), adminHandler.DeleteSong)
	adminOnly.Patch("/songs/:id/status", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.ToggleSongStatus)

	// SongVariant Operations
	adminOnly.Get("/variants", adminHandler.GetVariants)
	adminOnly.Get("/videos", adminHandler.GetVideos)
	adminOnly.Get("/variants/:id<int>", adminHandler.GetVariant)
	adminOnly.Post("/variants", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.CreateVariant)
	adminOnly.Put("/variants/:id<int>", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.UpdateVariant)
	adminOnly.Put("/variants/:id<int>/video", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.UpdateVariantVideo)
	adminOnly.Patch("/variants/:id<int>/status", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.ToggleVariantStatus)
	adminOnly.Patch("/variants/:id<int>/spoiler", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.ToggleVariantSpoiler)
	adminOnly.Patch("/variants/:id<int>/nsfw", middleware.HasPermissionMiddleware("song.edit", userRepo), adminHandler.ToggleVariantNSFW)
	adminOnly.Delete("/variants/:id<int>", middleware.HasPermissionMiddleware("song.delete", userRepo), adminHandler.DeleteVariant)

	// Artist Operations
	adminOnly.Get("/artists", adminHandler.GetArtists)
	adminOnly.Get("/artists/:id", adminHandler.GetArtist)
	adminOnly.Patch("/artists/:id/status", middleware.HasPermissionMiddleware("artist.edit", userRepo), adminHandler.ToggleArtistStatus)
	adminOnly.Post("/artists", middleware.HasPermissionMiddleware("artist.create", userRepo), adminHandler.CreateArtist)
	adminOnly.Put("/artists/:id", middleware.HasPermissionMiddleware("artist.edit", userRepo), adminHandler.UpdateArtist)
	adminOnly.Post("/artists/:id/avatar/generate", adminHandler.GenerateArtistAvatar)
	adminOnly.Post("/artists/:id/sync-avatar", adminHandler.SyncArtistAvatar)
	adminOnly.Post("/artists/generate-avatars", adminHandler.BatchGenerateArtistAvatars)
	adminOnly.Post("/artists/merge", adminHandler.MergeArtists)
	adminOnly.Post("/artists/recount-songs", adminHandler.RecountArtistStats)
	adminOnly.Delete("/artists/:id", adminHandler.DeleteArtist)

	// Taxonomy Operations
	adminOnly.Get("/years", adminHandler.GetYears)
	adminOnly.Get("/seasons", adminHandler.GetSeasons)
	adminOnly.Get("/formats", adminHandler.GetFormats)
	adminOnly.Get("/song-types", adminHandler.GetSongTypes)

	// Years
	adminOnly.Get("/taxonomies/years", adminHandler.GetYears)
	adminOnly.Post("/taxonomies/years", middleware.HasPermissionMiddleware("taxonomy.years.create", userRepo), adminHandler.CreateYear)
	adminOnly.Put("/taxonomies/years/:id", middleware.HasPermissionMiddleware("taxonomy.years.edit", userRepo), adminHandler.UpdateYear)
	adminOnly.Patch("/taxonomies/years/:id/current", middleware.HasPermissionMiddleware("taxonomy.years.edit", userRepo), adminHandler.ToggleYearCurrent)
	adminOnly.Delete("/taxonomies/years/:id", middleware.HasPermissionMiddleware("taxonomy.years.delete", userRepo), adminHandler.DeleteYear)

	// Seasons
	adminOnly.Get("/taxonomies/seasons", adminHandler.GetSeasons)
	adminOnly.Post("/taxonomies/seasons", middleware.HasPermissionMiddleware("taxonomy.seasons.create", userRepo), adminHandler.CreateSeason)
	adminOnly.Put("/taxonomies/seasons/:id", middleware.HasPermissionMiddleware("taxonomy.seasons.edit", userRepo), adminHandler.UpdateSeason)
	adminOnly.Patch("/taxonomies/seasons/:id/current", middleware.HasPermissionMiddleware("taxonomy.seasons.edit", userRepo), adminHandler.ToggleSeasonCurrent)
	adminOnly.Delete("/taxonomies/seasons/:id", middleware.HasPermissionMiddleware("taxonomy.seasons.delete", userRepo), adminHandler.DeleteSeason)

	// Formats
	adminOnly.Get("/taxonomies/formats", adminHandler.GetFormats)
	adminOnly.Post("/taxonomies/formats", middleware.HasPermissionMiddleware("taxonomy.formats.create", userRepo), adminHandler.CreateFormat)
	adminOnly.Put("/taxonomies/formats/:id", middleware.HasPermissionMiddleware("taxonomy.formats.edit", userRepo), adminHandler.UpdateFormat)
	adminOnly.Delete("/taxonomies/formats/:id", middleware.HasPermissionMiddleware("taxonomy.formats.delete", userRepo), adminHandler.DeleteFormat)

	// Genres
	adminOnly.Get("/genres", adminHandler.GetGenres)
	adminOnly.Post("/taxonomies/genres", middleware.HasPermissionMiddleware("taxonomy.genres.create", userRepo), adminHandler.CreateGenre)
	adminOnly.Put("/taxonomies/genres/:id", middleware.HasPermissionMiddleware("taxonomy.genres.edit", userRepo), adminHandler.UpdateGenre)
	adminOnly.Delete("/taxonomies/genres/:id", middleware.HasPermissionMiddleware("taxonomy.genres.delete", userRepo), adminHandler.DeleteGenre)

	// Studios & Producers (Search/List)
	adminOnly.Get("/studios", adminHandler.GetStudios)
	adminOnly.Get("/producers", adminHandler.GetProducers)

	// Badges
	adminOnly.Get("/badges", badgeHandler.GetAll)
	adminOnly.Get("/badges/:id", badgeHandler.GetByID)
	adminOnly.Post("/badges", middleware.HasPermissionMiddleware("badge.manage", userRepo), badgeHandler.Create)
	adminOnly.Put("/badges/:id", middleware.HasPermissionMiddleware("badge.manage", userRepo), badgeHandler.Update)
	adminOnly.Delete("/badges/:id", middleware.HasPermissionMiddleware("badge.manage", userRepo), badgeHandler.Delete)

	// Tournament Operations
	adminOnly.Get("/tournaments", tournamentHandler.ListTournaments)
	adminOnly.Get("/tournaments/:id<int>", tournamentHandler.GetTournament)
	adminOnly.Post("/tournaments", middleware.HasPermissionMiddleware("tournament.manage", userRepo), tournamentHandler.CreateTournament)
	adminOnly.Post("/tournaments/:id<int>/seed", middleware.HasPermissionMiddleware("tournament.manage", userRepo), tournamentHandler.SeedTournament)
	adminOnly.Post("/tournaments/:id<int>/advance", middleware.HasPermissionMiddleware("tournament.manage", userRepo), tournamentHandler.AdvanceTournament)
	adminOnly.Delete("/tournaments/:id<int>", middleware.HasPermissionMiddleware("tournament.manage", userRepo), tournamentHandler.DeleteTournament)

	// Announcement Operations
	adminOnly.Get("/announcements", announcementHandler.GetAllAnnouncements)
	adminOnly.Get("/announcements/:id", announcementHandler.GetAnnouncementByID)
	adminOnly.Post("/announcements", middleware.HasPermissionMiddleware("announcement.manage", userRepo), announcementHandler.CreateAnnouncement)
	adminOnly.Put("/announcements/:id", middleware.HasPermissionMiddleware("announcement.manage", userRepo), announcementHandler.UpdateAnnouncement)
	adminOnly.Delete("/announcements/:id", middleware.HasPermissionMiddleware("announcement.manage", userRepo), announcementHandler.DeleteAnnouncement)
	adminOnly.Patch("/announcements/:id/toggle", middleware.HasPermissionMiddleware("announcement.manage", userRepo), announcementHandler.ToggleActive)
	adminOnly.Delete("/announcements/:id", middleware.HasPermissionMiddleware("announcement.manage", userRepo), announcementHandler.DeleteAnnouncement)

	// Permission Management (Owner/Admin only)
	adminOnly.Get("/permissions", permissionHandler.GetAllPermissions)
	adminOnly.Get("/roles/permissions", permissionHandler.GetRoles)
	adminOnly.Post("/roles/:id/permissions", middleware.AdminMiddleware(), permissionHandler.UpdateRolePermissions)
}
