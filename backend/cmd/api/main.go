package main

import (
	"context"
	"log"
	"os"

	"anirank/api/internal/delivery/http"
	"anirank/api/internal/delivery/http/middleware"
	v1 "anirank/api/internal/delivery/http/v1"
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/google"
	"anirank/api/internal/infrastructure/og"
	"anirank/api/internal/jobs"
	"anirank/api/internal/repository/postgres"
	"anirank/api/internal/usecase"
	"anirank/api/internal/usecase/admin"
	"anirank/api/internal/usecase/audit"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/interaction"
	"anirank/api/internal/usecase/moderation"
	"anirank/api/internal/usecase/playlist"
	"anirank/api/internal/usecase/public"
	"anirank/api/internal/usecase/tournament"

	"anirank/api/internal/infrastructure/cache"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

// @title Anirank API
// @version 1.0
// @description API Rest interactiva y pública para el proyecto Anirank.
// @termsOfService http://swagger.io/terms/

// @contact.name Soporte API Anirank
// @contact.email admin@anirank.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	// Load env: repo-root .env (OAuth keys, APP_URL) then cwd .env (backend overrides e.g. DB_HOST=db).
	_ = godotenv.Load("../.env")
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found in cwd, using OS env vars")
	}

	appPort := os.Getenv("APP_PORT")
	if appPort == "" {
		appPort = "3000"
	}

	// Setup DB connection
	dbURL := os.Getenv("DATABASE_URL")
	var db *sqlx.DB
	var err error

	if dbURL != "" {
		db, err = infrastructure.NewDatabaseConnectionFromURL(dbURL)
	} else {
		dbUser := os.Getenv("DB_USER")
		dbPass := os.Getenv("DB_PASSWORD")
		dbHost := os.Getenv("DB_HOST")
		dbPort := os.Getenv("DB_PORT")
		dbName := os.Getenv("DB_NAME")
		db, err = infrastructure.NewDatabaseConnection(dbUser, dbPass, dbHost, dbPort, dbName)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// 1. Dependency Injection: Repositories
	animeRepo := postgres.NewAnimeRepository(db)
	taxonomyRepo := postgres.NewTaxonomyRepository(db)
	userRepo := postgres.NewUserRepository(db)
	interactionRepo := postgres.NewInteractionRepository(db)
	commentRepo := postgres.NewCommentRepository(db)
	playlistRepo := postgres.NewPlaylistRepository(db)
	songRepo := postgres.NewSongRepository(db)
	variantRepo := postgres.NewSongVariantRepository(db)
	artistRepo := postgres.NewArtistRepository(db)
	moderationRepo := postgres.NewModerationRepository(db)
	tournamentRepo := postgres.NewTournamentRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	auditRepo := postgres.NewAuditLogRepository(db)
	jobsRepo := postgres.NewJobsRepository(db)
	adminRepo := postgres.NewAdminRepository(db)
	xpRepo := postgres.NewXPRepository(db)

	// Setup S3 Storage
	s3Access := os.Getenv("S3_ACCESS_KEY")
	s3Secret := os.Getenv("S3_SECRET_KEY")
	s3Region := os.Getenv("S3_REGION")
	s3Bucket := os.Getenv("S3_BUCKET")
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	s3PublicUrl := os.Getenv("S3_PUBLIC_URL")

	// Ensure endpoint has protocol if not set natively
	if s3Endpoint != "" && len(s3Endpoint) > 4 && s3Endpoint[:4] != "http" {
		s3Endpoint = "http://" + s3Endpoint
	}

	storageService, err := infrastructure.NewS3Storage(context.Background(), s3Access, s3Secret, s3Region, s3Bucket, s3Endpoint, s3PublicUrl)
	if err != nil {
		log.Printf("Warning: Could not initialize S3 Storage: %v", err)
	}

	mediaService := infrastructure.NewMediaService(storageService)

	// Setup Cache (Optional Redis)
	redisURL := os.Getenv("REDIS_URL")
	var appCache domain.Cache
	if redisURL != "" {
		rc, err := cache.NewRedisCache(redisURL)
		if err != nil {
			log.Printf("Warning: Redis requested but connection failed: %v. Falling back to NoOpCache.", err)
			appCache = cache.NewNoOpCache()
		} else {
			log.Println("✅ Redis cache connected successfully")
			appCache = rc
		}
	} else {
		log.Println("Note: Redis not configured, using NoOpCache (caching disabled)")
		appCache = cache.NewNoOpCache()
	}

	// 2. Dependency Injection: Services & Usecases
	jwtService := auth.NewJWTService()

	anilistClient := anilist.NewClient()
	googleClient := google.NewClient()

	discoveryUsecase := public.NewDiscoveryUsecase(taxonomyRepo)
	animeUsecase := public.NewAnimeUsecase(animeRepo, songRepo, mediaService)
	searchUsecase := public.NewSearchUsecase(animeRepo, songRepo, artistRepo, taxonomyRepo, userRepo, storageService)
	catalogUsecase := public.NewCatalogUsecase(animeRepo, songRepo, artistRepo, taxonomyRepo, userRepo, playlistRepo, interactionRepo, moderationRepo, mediaService, appCache)
	xpUsecase := usecase.NewXPUsecase(xpRepo, userRepo)
	activityUsecase := usecase.NewActivityUsecase(postgres.NewActivityRepository(db), userRepo, songRepo, artistRepo, mediaService)
	authUsecase := auth.NewAuthUsecase(userRepo, jwtService, storageService, mediaService, xpUsecase, anilistClient, googleClient, os.Getenv("ENCRYPTION_KEY"))
	interactionUsecase := interaction.NewInteractionUsecase(interactionRepo, commentRepo, userRepo, notificationRepo, songRepo, animeRepo, mediaService, xpUsecase, activityUsecase)
	playlistUsecase := playlist.NewPlaylistUsecase(playlistRepo, songRepo, animeRepo, interactionRepo, mediaService, xpUsecase, userRepo)
	auditUsecase := audit.NewAuditLogUsecase(auditRepo)

	userAdminUsecase := admin.NewUserAdminUsecase(userRepo, mediaService, auditUsecase)
	contentAdminUsecase := admin.NewContentAdminUsecase(animeRepo, songRepo, variantRepo, artistRepo, taxonomyRepo, userRepo, anilistClient, mediaService, auditUsecase)
	adminUsecase := admin.NewAdminUsecase(userAdminUsecase, contentAdminUsecase, adminRepo, moderationRepo, jobsRepo)

	ogGenerator := og.NewGenerator(s3PublicUrl, s3Endpoint)
	shareHandler := v1.NewShareHandler(animeUsecase, catalogUsecase, playlistUsecase)

	seoUsecase := public.NewSEOUsecase(animeRepo, songRepo, artistRepo, userRepo, playlistRepo, ogGenerator.GetVersion)
	seoHandler := v1.NewSEOHandler(seoUsecase)

	moderationUsecase := moderation.NewModerationUsecase(moderationRepo, notificationRepo, mediaService)
	tournamentUsecase := tournament.NewTournamentUsecase(tournamentRepo, songRepo, animeRepo, storageService)

	// --- 1.5 Start Background Cron Scheduler ---
	cronInstance := jobs.StartCronScheduler(jobsRepo, tournamentUsecase)
	defer cronInstance.Stop()

	// 3. Setup Fiber Framework
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	// CORS — allow the SPA origin through
	allowOrigins := os.Getenv("ALLOW_ORIGINS")
	if allowOrigins == "" {
		allowOrigins = "http://localhost:5173, http://localhost:4173, http://localhost:8080, https://anirank.work"
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowOrigins,
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
	}))

	app.Use(middleware.RequestLogger())

	// Setup Daily Stats
	statsRepo := postgres.NewStatsRepository(db)
	statsUsecase := public.NewStatsUsecase(statsRepo, appCache)

	// 4. Register Routes
	http.SetupPublicRoutes(app, db, discoveryUsecase, animeUsecase, searchUsecase, catalogUsecase, authUsecase, interactionUsecase, playlistUsecase, adminUsecase, moderationUsecase, tournamentUsecase, auditUsecase, jwtService, storageService, mediaService, xpUsecase, activityUsecase, statsUsecase, ogGenerator, shareHandler, seoHandler)

	// Run Server
	log.Printf("Starting server on port %s...", appPort)
	if err := app.Listen(":" + appPort); err != nil {
		log.Fatalf("Server fell over: %v", err)
	}
}
