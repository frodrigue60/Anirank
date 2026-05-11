package main

import (
	"anirank/api/internal/delivery/http"
	"anirank/api/internal/delivery/http/middleware"
	v1 "anirank/api/internal/delivery/http/v1"
	"anirank/api/internal/domain"
	"anirank/api/internal/infrastructure"
	"anirank/api/internal/infrastructure/anilist"
	"anirank/api/internal/infrastructure/discord"
	"anirank/api/internal/infrastructure/google"
	"anirank/api/internal/infrastructure/mail"
	"anirank/api/internal/infrastructure/og"
	"anirank/api/internal/jobs"
	"anirank/api/internal/repository/postgres"
	"anirank/api/internal/usecase"
	"anirank/api/internal/usecase/admin"
	"anirank/api/internal/usecase/audit"
	"anirank/api/internal/usecase/auth"
	"anirank/api/internal/usecase/interaction"
	"anirank/api/internal/usecase/moderation"
	"anirank/api/internal/usecase/notification"
	"anirank/api/internal/usecase/playlist"
	"anirank/api/internal/usecase/public"
	"anirank/api/internal/usecase/tournament"
	"context"
	"fmt"
	"log"
	"os"

	"anirank/api/internal/infrastructure/cache"

	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/storage/memory/v2"
	"github.com/gofiber/storage/redis/v3"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	goredis "github.com/redis/go-redis/v9"
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
	// Priority 1: Connection URLs (Railway standard)
	dbURL := getEnvWithFallback("DATABASE_URL", "DATABASE_PUBLIC_URL", "MYSQL_URL", "MYSQL_PRIVATE_URL")

	var db *sqlx.DB
	var err error

	if dbURL != "" {
		db, err = infrastructure.NewDatabaseConnectionFromURL(dbURL)
	} else {
		// Priority 2: Inferred driver from specific host variables
		driver := os.Getenv("DB_DRIVER")
		if driver == "" {
			if os.Getenv("PGHOST") != "" || os.Getenv("POSTGRES_USER") != "" {
				driver = "postgres"
			} else if os.Getenv("MYSQLHOST") != "" || os.Getenv("MYSQLUSER") != "" {
				driver = "mysql"
			} else {
				driver = "postgres" // default
			}
		}

		// Support all Railway/Standard naming conventions (12-factor fallback)
		dbUser := getEnvWithFallback("DB_USER", "PGUSER", "POSTGRES_USER", "MYSQLUSER", "MYSQL_USER")
		dbPass := getEnvWithFallback("DB_PASSWORD", "PGPASSWORD", "POSTGRES_PASSWORD", "MYSQLPASSWORD", "MYSQL_PASSWORD")
		dbHost := getEnvWithFallback("DB_HOST", "PGHOST", "POSTGRES_HOST", "MYSQLHOST", "MYSQL_HOST")
		dbPort := getEnvWithFallback("DB_PORT", "PGPORT", "POSTGRES_PORT", "MYSQLPORT", "MYSQL_PORT")
		dbName := getEnvWithFallback("DB_NAME", "PGDATABASE", "POSTGRES_DB", "MYSQLDATABASE", "MYSQL_DATABASE")

		db, err = infrastructure.NewDatabaseConnection(driver, dbUser, dbPass, dbHost, dbPort, dbName)
	}

	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := infrastructure.RunMigrations(db, "./database/migrations"); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
		// We don't necessarily want to fatal here if the DB is already mostly correct,
		// but for production it might be safer to stop.
		// For now, let's just log it to avoid blocking startup if it's a minor issue.
	}

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
	searchRepo := postgres.NewSearchRepository(db)

	// Seed base data (Score Formats, Song Types)
	sfSeeder := postgres.NewScoreFormatSeeder(db)
	if err := sfSeeder.Seed(context.Background()); err != nil {
		log.Printf("Warning: Failed to seed score formats: %v", err)
	}

	stSeeder := postgres.NewSongTypeSeeder(db)
	if err := stSeeder.Seed(context.Background()); err != nil {
		log.Printf("Warning: Failed to seed song types: %v", err)
	}

	storageService, err := infrastructure.InitStorageFromEnv(context.Background())
	if err != nil {
		log.Printf("Warning: Could not initialize Storage Service: %v", err)
	}

	mediaService := infrastructure.NewMediaService(storageService)

	// Setup Cache (Optional Redis)
	redisEnabled := os.Getenv("REDIS_ENABLED") != "false"

	// Resolve Redis URL (Railway/Production Compatibility)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		// Try to construct from individual components
		host := os.Getenv("REDIS_HOST")
		if host == "" {
			host = os.Getenv("REDISHOST") // Railway format
		}

		if host != "" {
			port := os.Getenv("REDIS_PORT")
			if port == "" {
				port = os.Getenv("REDISPORT")
			}
			if port == "" {
				port = "6379"
			}

			user := os.Getenv("REDIS_USER")
			if user == "" {
				user = os.Getenv("REDISUSER")
			}

			pass := os.Getenv("REDIS_PASSWORD")
			if pass == "" {
				pass = os.Getenv("REDISPASSWORD")
			}

			db := os.Getenv("REDIS_DB")
			if db == "" {
				db = "0"
			}

			if pass != "" {
				redisURL = fmt.Sprintf("redis://%s:%s@%s:%s/%s", user, pass, host, port, db)
			} else {
				redisURL = fmt.Sprintf("redis://%s:%s/%s", host, port, db)
			}
		}
	}

	var appCache domain.Cache
	if redisURL != "" && redisEnabled {
		rc, err := cache.NewRedisCache(redisURL)
		if err != nil {
			log.Printf("Warning: Redis requested but connection failed: %v. Falling back to NoOpCache.", err)
			appCache = cache.NewNoOpCache()
		} else {
			log.Println("✅ Redis cache connected successfully")
			appCache = rc
		}
	} else {
		if !redisEnabled && redisURL != "" {
			log.Println("ℹ️ Redis is explicitly disabled via REDIS_ENABLED=false")
		} else {
			log.Println("Note: Redis not configured, using NoOpCache (caching disabled)")
		}
		appCache = cache.NewNoOpCache()
	}

	// --- 1.6 Setup Rate Limiting Storage (Resilience Shield) ---
	var limitStorage fiber.Storage
	if redisURL != "" && redisEnabled {
		// We use a recovery block because fiber/storage/redis/v3 panics if it can't connect/resolve at startup
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Rate Limiter: Redis storage initialization panicked: %v. Falling back to Memory storage.", r)
					limitStorage = memory.New()
				}
			}()

			// Use NewFromConnection to apply custom timeouts (fixes the 5s DNS timeout block)
			opts, err := goredis.ParseURL(redisURL)
			if err != nil {
				log.Printf("⚠️ Rate Limiter: Invalid Redis URL: %v", err)
				return
			}
			opts.DialTimeout = 1 * time.Second
			opts.ReadTimeout = 1 * time.Second
			opts.WriteTimeout = 1 * time.Second

			client := goredis.NewClient(opts)
			primaryStorage := redis.NewFromConnection(client)

			limitStorage = cache.NewResilientStorage(primaryStorage)
			log.Println("✅ Rate Limiter: Resilient Redis storage initialized with custom timeouts")
		}()
	}

	if limitStorage == nil {
		limitStorage = memory.New()
		log.Println("⚠️ Rate Limiter: Using Memory storage (Redis unavailable or not configured)")
	}

	// 2. Dependency Injection: Services & Usecases
	jwtService := auth.NewJWTService()

	anilistClient := anilist.NewClient()
	googleClient := google.NewClient()
	discordClient := discord.NewClient()

	discoveryUsecase := public.NewDiscoveryUsecase(taxonomyRepo, songRepo)
	animeUsecase := public.NewAnimeUsecase(animeRepo, songRepo, mediaService)
	searchUsecase := public.NewSearchUsecase(searchRepo, storageService)
	catalogUsecase := public.NewCatalogUsecase(animeRepo, songRepo, artistRepo, taxonomyRepo, userRepo, playlistRepo, interactionRepo, moderationRepo, anilistClient, mediaService, appCache, os.Getenv("ENCRYPTION_KEY"))

	badgeRepo := postgres.NewBadgeRepository(db)
	auditUsecase := audit.NewAuditLogUsecase(auditRepo)
	badgeUsecase := admin.NewBadgeUsecase(badgeRepo, userRepo, interactionRepo, commentRepo, storageService, auditUsecase)
	notificationUsecase := notification.NewNotificationUsecase(notificationRepo, appCache)

	xpUsecase := usecase.NewXPUsecase(xpRepo, userRepo, badgeUsecase)
	activityUsecase := usecase.NewActivityUsecase(postgres.NewActivityRepository(db), userRepo, songRepo, artistRepo, mediaService)

	// Auth specialized services
	tokenRepo := postgres.NewAuthTokenRepository(db)
	resendAPIKey := os.Getenv("RESEND_API_KEY")
	resendFrom := os.Getenv("RESEND_FROM_EMAIL")
	if resendFrom == "" {
		resendFrom = "AniRank <noreply@anirank.com>"
	}
	appURL := os.Getenv("APP_URL")
	if appURL == "" {
		appURL = "http://localhost:5173" // Default frontend dev URL
		log.Println("⚠️ APP_URL not set, defaulting to http://localhost:5173 for email links")
	} else {
		log.Printf("📧 Email service initialized with APP_URL: %s", appURL)
	}
	mailService := mail.NewResendService(resendAPIKey, resendFrom, appURL)

	authUsecase := auth.NewAuthUsecase(userRepo, jwtService, storageService, mediaService, xpUsecase, badgeUsecase, anilistClient, googleClient, discordClient, mailService, tokenRepo, os.Getenv("ENCRYPTION_KEY"))
	interactionUsecase := interaction.NewInteractionUsecase(interactionRepo, commentRepo, userRepo, notificationUsecase, songRepo, animeRepo, artistRepo, mediaService, xpUsecase, activityUsecase, badgeUsecase)
	playlistUsecase := playlist.NewPlaylistUsecase(playlistRepo, songRepo, animeRepo, interactionRepo, mediaService, xpUsecase, userRepo)

	userAdminUsecase := admin.NewUserAdminUsecase(userRepo, mediaService, auditUsecase)
	contentAdminUsecase := admin.NewContentAdminUsecase(animeRepo, songRepo, variantRepo, artistRepo, taxonomyRepo, userRepo, anilistClient, mediaService, auditUsecase, interactionRepo, notificationUsecase)
	adminUsecase := admin.NewAdminUsecase(userAdminUsecase, contentAdminUsecase, adminRepo, moderationRepo, jobsRepo)

	ogGenerator := og.NewGenerator(storageService.GetPublicURL(), storageService.GetEndpoint())
	shareHandler := v1.NewShareHandler(animeUsecase, catalogUsecase, playlistUsecase)

	seoUsecase := public.NewSEOUsecase(animeRepo, songRepo, artistRepo, userRepo, playlistRepo, ogGenerator.GetVersion)
	seoHandler := v1.NewSEOHandler(seoUsecase)

	moderationUsecase := moderation.NewModerationUsecase(moderationRepo, userRepo, notificationUsecase, mediaService)
	tournamentUsecase := tournament.NewTournamentUsecase(tournamentRepo, songRepo, animeRepo, storageService)

	// --- 1.5 Start Background Cron Scheduler ---
	cronInstance := jobs.StartCronScheduler(jobsRepo, tournamentUsecase)
	defer cronInstance.Stop()

	// 3. Setup Fiber Framework
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
		BodyLimit:    10 * 1024 * 1024, // 10MB global limit to let handlers manage specific limits
	})

	// --- Security & Middleware ---
	securityEnabled := os.Getenv("SECURITY_HEADERS_ENABLED") != "false"
	corsEnabled := os.Getenv("CORS_ENABLED") != "false"
	originsFilterEnabled := os.Getenv("ALLOW_ORIGINS_FILTER") != "false"

	if corsEnabled {
		allowOrigins := os.Getenv("ALLOW_ORIGINS")
		allowCredentials := true

		if !originsFilterEnabled {
			allowOrigins = "*"
			allowCredentials = false // Fiber panics if Credentials=true with wildcard origins
		} else if allowOrigins == "" {
			allowOrigins = "http://localhost:5173, http://localhost:4173, http://localhost:8080, https://anirank.work"
		}

		app.Use(cors.New(cors.Config{
			AllowOrigins:     allowOrigins,
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token",
			AllowMethods:     "GET, POST, PUT, PATCH, DELETE, OPTIONS",
			AllowCredentials: allowCredentials,
		}))
		log.Printf("🛡️  CORS: Enabled (Origins: %s, Credentials: %v)", allowOrigins, allowCredentials)
	} else {
		log.Println("⚠️  CORS: Disabled (Middleware skipped)")
	}

	app.Use(middleware.RequestLogger())

	if securityEnabled {
		app.Use(middleware.SecurityHeaders())
		log.Println("🛡️  Security Headers: Enabled")
	} else {
		log.Println("⚠️  Security Headers: Disabled (Middleware skipped)")
	}

	// Setup Daily Stats
	statsRepo := postgres.NewStatsRepository(db)
	statsUsecase := public.NewStatsUsecase(statsRepo, appCache)

	// 4. Register Routes
	http.SetupPublicRoutes(app, db, discoveryUsecase, animeUsecase, searchUsecase, catalogUsecase, authUsecase, interactionUsecase, playlistUsecase, adminUsecase, moderationUsecase, tournamentUsecase, auditUsecase, jwtService, storageService, mediaService, xpUsecase, activityUsecase, statsUsecase, ogGenerator, shareHandler, seoHandler, limitStorage, appCache)

	// Run Server
	log.Printf("Starting server on port %s...", appPort)
	if err := app.Listen(":" + appPort); err != nil {
		log.Fatalf("Server fell over: %v", err)
	}
}

func getEnvWithFallback(keys ...string) string {
	for _, key := range keys {
		val := os.Getenv(key)
		if val != "" {
			return val
		}
	}
	return ""
}
