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
	seoHandler *v1.SEOHandler) {

	// Core Repositories
	songRepo := postgres.NewSongRepository(db)
	userRepo := postgres.NewUserRepository(db)
	animeRepo := postgres.NewAnimeRepository(db)
	artistRepo := postgres.NewArtistRepository(db)
	playlistRepo := postgres.NewPlaylistRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	interactionRepo := postgres.NewInteractionRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)

	// HTTP Handlers
	discoveryHandler := v1.NewDiscoveryHandler(discoveryUsecase)
	animeHandler := v1.NewAnimeHandler(animeUsecase)
	searchHandler := v1.NewSearchHandler(searchUsecase)
	catalogHandler := v1.NewCatalogHandler(catalogUsecase)
	authHandler := v1.NewAuthHandler(authUsecase)
	interactionUsecase = interaction.NewInteractionUsecase(interactionRepo, commentRepo, userRepo, notificationRepo, songRepo, animeRepo, mediaService, xpUsecase, activityUsecase)
	interactionHandler := v1.NewInteractionHandler(interactionUsecase, activityUsecase, songRepo, userRepo, animeRepo, artistRepo, commentRepo)
	playlistHandler := v1.NewPlaylistHandler(playlistUsecase, playlistRepo, songRepo, userRepo)
	adminHandler := v1.NewAdminHandler(adminUsecase, songRepo, userRepo, animeRepo, artistRepo, playlistRepo)
	moderationHandler := v1.NewModerationHandler(moderationUsecase, songRepo, commentRepo)
	tournamentHandler := v1.NewTournamentHandler(tournamentUsecase)

	announcementRepo := postgres.NewAnnouncementRepository(db)
	announcementUsecase := announcement.NewAnnouncementUsecase(announcementRepo, storageService)
	announcementHandler := v1.NewAnnouncementHandler(announcementUsecase)

	badgeRepo := postgres.NewBadgeRepository(db)
	badgeUsecase := admin.NewBadgeUsecase(badgeRepo, storageService, auditLogUsecase)
	badgeHandler := v1.NewBadgeHandler(badgeUsecase)

	activityHandler := v1.NewActivityHandler(activityUsecase)

	notificationUsecase := notification.NewNotificationUsecase(notificationRepo)
	notificationHandler := v1.NewNotificationHandler(notificationUsecase)

	statsHandler := v1.NewStatsHandler(statsUsecase)
	ogHandler := v1.NewOGHandler(ogGenerator, animeUsecase, catalogUsecase, playlistUsecase, statsUsecase)

	// API V1 Group
	api := app.Group("/api")

	// --- SWAGGER DOCS ---
	api.Get("/swagger/*", swagger.HandlerDefault)

	// --- PUBLIC ROUTES ---

	// Auth
	api.Post("/login", authHandler.Login)
	api.Post("/register", authHandler.Register)
	api.Get("/auth/google/login", authHandler.GoogleLogin)
	api.Post("/auth/google/login-callback", authHandler.GoogleLoginCallback)
	api.Get("/auth/anilist/login", authHandler.AnilistLogin)
	api.Post("/auth/anilist/login-callback", authHandler.AnilistLoginCallback)

	// Init data for SPA
	api.Get("/init", discoveryHandler.Init)

	// Site Statistics
	api.Get("/site-statistics", statsHandler.GetSiteStats)

	// Global Search Engine
	api.Get("/search", searchHandler.Search)

	// Anime Endpoints
	api.Get("/animes", animeHandler.Index)
	api.Get("/animes/:slug", animeHandler.Show)

	// Catalog: Songs
	api.Get("/songs", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.SongIndex)
	api.Get("/songs/ranking/:type", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.SongRanking)
	api.Get("/songs/:uuid/comments", middleware.OptionalAuthMiddleware(jwtService), interactionHandler.GetSongComments)
	api.Get("/songs/:anime_slug/:song_slug", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.SongShow)
	api.Get("/animes/:anime_slug/songs/:song_slug", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.SongShow)

	// Catalog: Artists
	api.Get("/artists", catalogHandler.ArtistIndex)
	api.Get("/artists/:slug", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.ArtistShow)
	api.Get("/artists/:slug/songs", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.ArtistShow)

	// Catalog: Studios
	api.Get("/studios", catalogHandler.StudioIndex)
	api.Get("/studios/:slug", catalogHandler.StudioShow)

	// Catalog: Producers
	api.Get("/producers", catalogHandler.ProducerIndex)
	api.Get("/producers/:slug", catalogHandler.ProducerShow)

	// Catalog: Users
	api.Get("/users/ranking", catalogHandler.UserRanking)
	api.Get("/users/:slug", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.UserProfile)
	// Catalog: Home
	api.Get("/home", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.Home)

	// Catalog: Playlists
	api.Get("/playlists", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.PlaylistIndex)

	// Sitemap
	api.Get("/catalog/sitemap", catalogHandler.GetSitemap)
	api.Get("/catalog/sitemap.xml", catalogHandler.GetSitemapXML)

	// Activity Feed
	api.Get("/activities", activityHandler.Index)
	api.Get("/activities/recent", activityHandler.Recent)

	// Announcements Public
	api.Get("/announcements", announcementHandler.GetPublicAnnouncements)

	// Tournaments Public
	api.Get("/tournaments", tournamentHandler.ListTournamentsPublic)
	api.Get("/tournaments/active", tournamentHandler.GetActiveTournament)
	api.Get("/tournaments/:slug", middleware.OptionalAuthMiddleware(jwtService), tournamentHandler.GetTournamentBySlug)

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
	api.Get("/comments", middleware.OptionalAuthMiddleware(jwtService), interactionHandler.GetComments)
	api.Get("/comments/:id/replies", middleware.OptionalAuthMiddleware(jwtService), interactionHandler.GetReplies)

	// User Public Playlists
	api.Get("/playlists/users/:id", middleware.OptionalAuthMiddleware(jwtService), playlistHandler.GetUserPlaylists)
	api.Get("/playlists/:id", middleware.OptionalAuthMiddleware(jwtService), playlistHandler.GetPlaylist)

	// User Public Data
	api.Get("/users/:slug/playlists", middleware.OptionalAuthMiddleware(jwtService), catalogHandler.UserPlaylists)
	api.Get("/users/:slug/followers", catalogHandler.UserFollowers)
	api.Get("/users/:slug/following", catalogHandler.UserFollowing)
	api.Post("/users/favorites/themes", catalogHandler.UserFavorites)
	api.Post("/users/favorites/artists", catalogHandler.UserArtistFavorites)

	// --- PROTECTED ROUTES ---
	protected := api.Group("/", middleware.AuthMiddleware(jwtService))

	// User Profile
	protected.Get("/profile", authHandler.Profile)
	protected.Post("/users/avatar", authHandler.UpdateAvatar)
	protected.Post("/users/banner", authHandler.UpdateBanner)
	protected.Post("/users/score-format", authHandler.UpdateScoreFormat)
	protected.Patch("/users/profile", authHandler.UpdateProfile)

	// Anilist Link
	protected.Get("/auth/anilist/link", authHandler.AnilistLink)
	protected.Post("/auth/anilist/callback", authHandler.AnilistCallback)

	// Google Link
	protected.Get("/auth/google/link", authHandler.GoogleLink)
	protected.Post("/auth/google/callback", authHandler.GoogleCallback)
	protected.Post("/interactions/ratings", interactionHandler.Rate)
	protected.Post("/interactions/reactions", interactionHandler.React)
	protected.Post("/interactions/favorites", interactionHandler.ToggleFavorite)
	protected.Post("/comments", interactionHandler.SongComment)
	protected.Delete("/comments/:id", interactionHandler.DeleteComment)

	protected.Post("/users/:id/follow", interactionHandler.FollowUser)
	protected.Delete("/users/:id/follow", interactionHandler.UnfollowUser)
	protected.Get("/users/:id/is-following", interactionHandler.IsFollowing)

	// Notifications
	protected.Get("/notifications", notificationHandler.Index)
	protected.Put("/notifications/:id/read", notificationHandler.MarkAsRead)
	protected.Post("/notifications/read-all", notificationHandler.MarkAllAsRead)
	protected.Delete("/notifications/:id", notificationHandler.Delete)
	protected.Get("/notifications/unread-count", notificationHandler.GetUnreadCount)

	// --- PROTECTED PLAYLISTS ---
	protected.Get("/me/playlists", playlistHandler.GetMyPlaylists)
	protected.Post("/playlists", playlistHandler.Create)
	protected.Put("/playlists/:id", playlistHandler.Update)
	protected.Delete("/playlists/:id", playlistHandler.Delete)
	protected.Post("/playlists/:id/songs", playlistHandler.AddSong)
	protected.Delete("/playlists/:id/songs/:songID", playlistHandler.RemoveSong)
	protected.Put("/playlists/:id/songs/reorder", playlistHandler.ReorderSongs)

	// Protected Moderation / User Support
	protected.Post("/songs/reports", moderationHandler.CreateSongReport)
	protected.Post("/comments/reports", moderationHandler.CreateCommentReport)
	protected.Post("/user-requests", moderationHandler.CreateUserRequest)

	// Protected Tournament Voting
	protected.Post("/tournaments/matchups/:id/vote", tournamentHandler.SubmitVote)

	// --- STAFF EXCLUSIVE ROUTES ---
	adminOnly := protected.Group("/admin", middleware.StaffMiddleware())

	// Dashboard & System
	adminOnly.Get("/dashboard", adminHandler.GetDashboard)
	adminOnly.Post("/og/flush", ogHandler.FlushOGCache)

	// Admin Moderation Tickets Review
	adminOnly.Get("/songs/reports", moderationHandler.GetSongReports)
	adminOnly.Get("/songs/reports/:id", moderationHandler.GetSongReport)
	adminOnly.Put("/songs/reports/:id/resolve", moderationHandler.ResolveSongReport)
	adminOnly.Delete("/songs/reports/:id", moderationHandler.DeleteSongReport)
	
	adminOnly.Get("/comments/reports", moderationHandler.GetCommentReports)
	adminOnly.Get("/comments/reports/:id", moderationHandler.GetCommentReport)
	adminOnly.Put("/comments/reports/:id/resolve", moderationHandler.ResolveCommentReport)
	adminOnly.Delete("/comments/reports/:id", moderationHandler.DeleteCommentReport)
	adminOnly.Get("/user-requests", moderationHandler.GetUserRequests)
	adminOnly.Get("/user-requests/:id", moderationHandler.GetUserRequest)
	adminOnly.Patch("/user-requests/:id/status", moderationHandler.UpdateUserRequestStatus)
	adminOnly.Delete("/user-requests/:id", moderationHandler.DeleteUserRequest)

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
	adminOnly.Post("/users", adminHandler.CreateUser)
	adminOnly.Get("/roles", adminHandler.GetRoles)
	adminOnly.Get("/users/:id", adminHandler.GetUser)
	adminOnly.Put("/users/:id", adminHandler.UpdateUser)
	adminOnly.Delete("/users/:id", adminHandler.DeleteUser)
	adminOnly.Post("/users/:id/reset-password", adminHandler.ResetPassword)

	// Anime Operations
	adminOnly.Get("/animes", adminHandler.GetAnimes)
	adminOnly.Get("/animes/anilist-search", adminHandler.SearchAnilist)
	adminOnly.Post("/animes/from-anilist", adminHandler.CreateAnimeFromAnilist)
	adminOnly.Post("/animes/batch-from-anilist", adminHandler.BatchCreateAnimesFromAnilist)
	adminOnly.Get("/animes/:id", adminHandler.GetAnime)
	adminOnly.Post("/animes", adminHandler.CreateAnime)
	adminOnly.Post("/animes/batch", adminHandler.BatchFetchAnimes)
	adminOnly.Post("/animes/hydrate", adminHandler.HydrateAnimeSeason)
	adminOnly.Put("/animes/:id", adminHandler.UpdateAnime)
	adminOnly.Patch("/animes/:id/status", adminHandler.ToggleAnimeStatus)
	adminOnly.Post("/animes/:id/sync", adminHandler.SyncAnime)
	adminOnly.Delete("/animes/:id", adminHandler.DeleteAnime)
	adminOnly.Post("/animes/batch-delete", adminHandler.BatchDeleteAnimes)

	// Songs Group
	adminOnly.Get("/songs", adminHandler.GetSongs)
	adminOnly.Get("/songs/latest-number", adminHandler.GetLatestSongNumber)
	adminOnly.Get("/songs/:id", adminHandler.GetSong)
	adminOnly.Post("/songs", adminHandler.CreateSong)
	adminOnly.Put("/songs/:id", adminHandler.UpdateSong)
	adminOnly.Delete("/songs/:id", adminHandler.DeleteSong)
	adminOnly.Patch("/songs/:id/status", adminHandler.ToggleSongStatus)

	// SongVariant Operations
	adminOnly.Get("/variants", adminHandler.GetVariants)
	adminOnly.Get("/videos", adminHandler.GetVideos)
	adminOnly.Get("/variants/:id<int>", adminHandler.GetVariant)
	adminOnly.Post("/variants", adminHandler.CreateVariant)
	adminOnly.Put("/variants/:id<int>", adminHandler.UpdateVariant)
	adminOnly.Put("/variants/:id<int>/video", adminHandler.UpdateVariantVideo)
	adminOnly.Patch("/variants/:id<int>/status", adminHandler.ToggleVariantStatus)
	adminOnly.Delete("/variants/:id<int>", adminHandler.DeleteVariant)

	// Artist Operations
	adminOnly.Get("/artists", adminHandler.GetArtists)
	adminOnly.Get("/artists/:id", adminHandler.GetArtist)
	adminOnly.Patch("/artists/:id/status", adminHandler.ToggleArtistStatus)
	adminOnly.Post("/artists", adminHandler.CreateArtist)
	adminOnly.Put("/artists/:id", adminHandler.UpdateArtist)
	adminOnly.Post("/artists/:id/avatar/generate", adminHandler.GenerateArtistAvatar)
	adminOnly.Post("/artists/generate-avatars", adminHandler.BatchGenerateArtistAvatars)
	adminOnly.Post("/artists/merge", adminHandler.MergeArtists)
	adminOnly.Post("/artists/recount-songs", adminHandler.RecountArtistStats)
	adminOnly.Delete("/artists/:id", adminHandler.DeleteArtist)

	// Taxonomy Operations
	adminOnly.Get("/years", adminHandler.GetYears)
	adminOnly.Get("/seasons", adminHandler.GetSeasons)
	adminOnly.Get("/formats", adminHandler.GetFormats)

	// Years
	adminOnly.Get("/taxonomies/years", adminHandler.GetYears)
	adminOnly.Post("/taxonomies/years", adminHandler.CreateYear)
	adminOnly.Put("/taxonomies/years/:id", adminHandler.UpdateYear)
	adminOnly.Patch("/taxonomies/years/:id/current", adminHandler.ToggleYearCurrent)
	adminOnly.Delete("/taxonomies/years/:id", adminHandler.DeleteYear)

	// Seasons
	adminOnly.Get("/taxonomies/seasons", adminHandler.GetSeasons)
	adminOnly.Post("/taxonomies/seasons", adminHandler.CreateSeason)
	adminOnly.Put("/taxonomies/seasons/:id", adminHandler.UpdateSeason)
	adminOnly.Patch("/taxonomies/seasons/:id/current", adminHandler.ToggleSeasonCurrent)
	adminOnly.Delete("/taxonomies/seasons/:id", adminHandler.DeleteSeason)

	// Formats
	adminOnly.Get("/taxonomies/formats", adminHandler.GetFormats)
	adminOnly.Post("/taxonomies/formats", adminHandler.CreateFormat)
	adminOnly.Put("/taxonomies/formats/:id", adminHandler.UpdateFormat)
	adminOnly.Delete("/taxonomies/formats/:id", adminHandler.DeleteFormat)

	// Genres
	adminOnly.Get("/genres", adminHandler.GetGenres)
	adminOnly.Post("/taxonomies/genres", adminHandler.CreateGenre)
	adminOnly.Put("/taxonomies/genres/:id", adminHandler.UpdateGenre)
	adminOnly.Delete("/taxonomies/genres/:id", adminHandler.DeleteGenre)

	// Studios & Producers (Search/List)
	adminOnly.Get("/studios", adminHandler.GetStudios)
	adminOnly.Get("/producers", adminHandler.GetProducers)

	// Badges
	adminOnly.Get("/badges", badgeHandler.GetAll)
	adminOnly.Get("/badges/:id", badgeHandler.GetByID)
	adminOnly.Post("/badges", badgeHandler.Create)
	adminOnly.Put("/badges/:id", badgeHandler.Update)
	adminOnly.Delete("/badges/:id", badgeHandler.Delete)

	// Tournament Operations
	adminOnly.Get("/tournaments", tournamentHandler.ListTournaments)
	adminOnly.Get("/tournaments/:id<int>", tournamentHandler.GetTournament)
	adminOnly.Post("/tournaments", tournamentHandler.CreateTournament)
	adminOnly.Post("/tournaments/:id<int>/seed", tournamentHandler.SeedTournament)
	adminOnly.Post("/tournaments/:id<int>/advance", tournamentHandler.AdvanceTournament)
	adminOnly.Delete("/tournaments/:id<int>", tournamentHandler.DeleteTournament)

	// Announcement Operations
	adminOnly.Get("/announcements", announcementHandler.GetAllAnnouncements)
	adminOnly.Get("/announcements/:id", announcementHandler.GetAnnouncementByID)
	adminOnly.Post("/announcements", announcementHandler.CreateAnnouncement)
	adminOnly.Put("/announcements/:id", announcementHandler.UpdateAnnouncement)
	adminOnly.Delete("/announcements/:id", announcementHandler.DeleteAnnouncement)
	adminOnly.Patch("/announcements/:id/toggle", announcementHandler.ToggleActive)
}
