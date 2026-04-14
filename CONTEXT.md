# CONTEXT.md — AniRank Architecture & Project Conventions

> This file documents the project architecture, business logic, and established conventions. It is the single source of truth for any agent or developer working on the backend.

---

## 1. Project Overview

**AniRank** is an anime music (openings/endings/inserts) discovery and ranking web application.

**Core Features:**
- Browse and search anime songs (openings, endings, inserts) with rich metadata.
- Rate, react to, and comment on songs.
- View weekly/seasonal rankings of anime themes.
- User profiles with XP/Level gamification, badges, followers, and playlists.
- AniList and Google OAuth account linking.
- Administrative backend for full content management.
- Public tournament system for community song voting.

---

## 2. Stack & Infrastructure

| Layer | Technology |
|-------|-----------|
| Frontend | SvelteKit 5, TailwindCSS v4, Bun |
| Backend | Go 1.25, GoFiber v2 |
| Database | PostgreSQL (via `sqlx` + raw SQL, NO ORM) |
| Cache | Redis (optional; falls back to `NoOpCache` if not configured) |
| Storage | S3-compatible (MinIO in Docker, AWS S3 in production) |
| Auth | JWT HS256 (`golang-jwt/jwt/v5`) |
| Scheduler | `robfig/cron/v3` for background jobs |
| Containerization | Docker Compose (separate `backend` and `frontend` services) |
| OG Images | Custom Go generator (`infrastructure/og`) |

**Env file loading order:** `./backend/.env` (primary configuration)

---

## 3. Backend Architecture: Clean Architecture (Layered)

The backend strictly follows **Clean Architecture**. Dependencies only flow inward:

```
[Delivery / HTTP] → [Usecase] → [Domain Interface] ← [Repository / Infrastructure]
```

### Layer Descriptions

| Layer | Location | Responsibility |
|-------|----------|---------------|
| **Domain** | `internal/domain/` | Pure business types, repository interfaces, domain errors. No external dependencies. |
| **Repository** | `internal/repository/postgres/` | Concrete DB implementations of domain interfaces. Raw SQL via `sqlx`. |
| **Usecase** | `internal/usecase/` | Business logic. Orchestrates repositories and infrastructure services. |
| **Delivery** | `internal/delivery/http/` | GoFiber route handlers. Parses requests, calls usecases, returns `DTO`-mapped responses. |
| **DTO** | `internal/dto/` | Data Transfer Objects and mapper functions. **The final sanitization filter before JSON output.** |
| **Infrastructure** | `internal/infrastructure/` | External services: S3, Redis, AniList API client, Google OAuth client, OG image generator. |

### Dependency Injection
All dependencies are wired manually in `cmd/api/main.go`. There is no IoC container. Constructors follow the pattern `New{Entity}Repository(db)` and `New{Entity}Usecase(deps...)`.

---

## 4. Identity & Security Model

### Hybrid ID Strategy
- **Internal (DB):** Sequential `uint64` integers as primary keys for maximum join performance.
- **External (API):** UUIDs (`string`) for all public-facing references. Slugs are used for human-readable URLs.

### ID Flow
```
DB: id (uint64, internal) ↔ uuid (string, external)
JWT: { user_uuid: string, roles: []string } — no numeric IDs
Middleware: looks up user by UUID, injects both user_id and user_uuid into fiber.Ctx.Locals
DTOs: map ID fields → UUID fields before JSON serialization
```

### JWT Claims Structure
```go
type Claims struct {
    UserUUID string   `json:"user_uuid"`
    Roles    []string `json:"roles"`
    jwt.RegisteredClaims
}
```
Token validity: **24 hours**. Signed with `JWT_SECRET` env var using HS256.

### Middleware Pipeline (ordered)
1. `cors.New(...)` — CORS headers
2. `middleware.RequestLogger()` — structured request logging
3. Route-level: `AuthMiddleware(jwtService, userRepo)` for protected routes
4. Route-level: `OptionalAuthMiddleware(jwtService, userRepo)` for semi-public routes (enriches context if token present, continues if not)
5. Group-level: `StaffMiddleware()` — requires `owner|admin|editor|creator` roles
6. Route-level: `AdminMiddleware()` — requires `owner|admin` roles
7. Route-level: `HasPermissionMiddleware("permission.slug", userRepo)` — fine-grained permission check

---

## 5. Domain Entities Summary

| Entity | Key Fields | Notes |
|--------|-----------|-------|
| `User` | `id`, `uuid`, `slug`, `email`, `xp`, `level` | Password hidden from JSON. Token/secret fields hidden. |
| `Anime` | `id`, `uuid`, `slug`, `title`, `format`, `status` | Has songs, genres, formats. |
| `Song` | `id`, `uuid`, `slug`, `type` (OP/ED/IN), `number` | Belongs to anime. Has artists via pivot. Has variants (video versions). |
| `SongVariant` | `id`, `uuid` | Video variant of a song (e.g. different quality/source). |
| `Artist` | `id`, `uuid`, `slug`, `name` | Linked to songs. Can have avatar. |
| `Playlist` | `id`, `uuid`, `slug` | User-created song collections. |
| `Badge` | `id`, `uuid`, `name`, `requirement_type`, `requirement_value` | Awarded automatically or manually. |
| `XPActivity` | `key` (string slug), `xp_amount`, `cooldown_seconds` | Defines how XP is awarded per action. |
| `Tournament` | `id`, `slug`, `matchups` | Community bracket voting system. |
| `Announcement` | `id`, `uuid`, `title`, `is_active` | Displayed in sidebar. |
| `Notification` | `id` (UUID used in responses), `user_id`, `type` | In-app notifications. |

---

## 6. RBAC (Role-Based Access Control)

Roles are stored in the `roles` table with a `weight` field (higher = more powerful).

| Role Slug | Access Level |
|-----------|-------------|
| `owner` | Full access, all permissions |
| `admin` | Admin panel, most permissions |
| `editor` | Can edit content (anime, songs, artists) |
| `creator` | Can create content |

Permissions are granular (e.g. `anime.create`, `song.edit`, `reports.manage`) and are stored in the `permissions` table, linked via `role_permissions`.

The `HasPermissionMiddleware` checks the authenticated user's permissions (loaded from DB) before allowing access to a specific route.

---

## 7. XP & Gamification System

XP is awarded via the `XPUsecase.AwardXP(ctx, userID, activityKey, metadata)` method.

`activityKey` values (defined in the `xp_activities` DB table, not in code):
- `daily_login`
- `song_rating`
- `song_comment`
- `playlist_created`
- (others as configured in DB)

Each activity has a `cooldown_seconds` to prevent spam. After XP is awarded, the system checks if the user leveled up and updates `users.level`.

Badges can also be auto-awarded based on `requirement_type` (e.g. `ratings_count >= 10`). This is handled by `BadgeUsecase.CheckAndAwardBadges()`.

---

## 8. Error Handling Convention

All handlers must return `domain.AppError`, not raw Go errors:

```go
// Correct
return domain.NewAppError(404, "Song not found", nil)
return domain.NewAppError(422, "Validation failed: email is required", nil)
return domain.NewAppError(500, "Database error", err) // err is logged, not exposed

// Wrong
return err // Raw errors leak implementation details and have no HTTP status code
```

The global `ErrorHandler` in `middleware/error_handler.go` intercepts `*domain.AppError` and formats it as:
```json
{ "success": false, "message": "...", "data": null }
```

---

## 9. DTO Layer Rules

- All mappers live in `internal/dto/mapper.go`.
- The `ID` field in any public-facing DTO **must be set to the UUID**, never the numeric ID.
- Token/secret fields (passwords, OAuth tokens) must use `json:"-"` in the domain struct and must NOT appear in any DTO.
- The canonical pattern for mapping a user is: `ToUserMinimalDTO` ⊂ `ToUserDTO` ⊂ `ToAuthUserDTO`.
- **Security Validation:** All mappers MUST be covered by unit tests in `mapper_test.go` using the `testutil.AssertNoInternalIDs` helper. This helper uses reflection to ensure no sequential IDs are leaked through the DTO hierarchy.

---

## 10. External Integrations

### AniList
- Client: `internal/infrastructure/anilist/`
- Used for: OAuth login/register, fetching user anime lists, syncing anime metadata.
- AniList user tokens are **encrypted** in the DB using the `ENCRYPTION_KEY` env var.

### Google OAuth
- Client: `internal/infrastructure/google/`
- Used for: OAuth login (auto-register if email is new), account linking for existing users.
- Redirect URIs are controlled by `GOOGLE_REDIRECT_URL` and `GOOGLE_LOGIN_REDIRECT_URL` env vars.
- **Note:** Google is responsible for sending password reset emails for Google-linked accounts. AniRank does not currently have its own SMTP server.

### S3 / MinIO
- Service: `internal/infrastructure/storage.go`
- Used for: Avatar, banner, badge icon, and artist avatar storage.
- `S3_PUBLIC_URL` controls the public base URL for all media.
- In Docker dev: MinIO runs at `localhost:9000`, public URL is typically `localhost:9000/{bucket}`.

---

## 11. Background Jobs

Managed by `robfig/cron/v3`. Defined in `internal/jobs/`.

| Job | Schedule | Purpose |
|-----|----------|---------|
| `TakeRankingSnapshot` | Configurable | Snapshots current song rating positions |
| `AdvanceTournament` | Configurable | Auto-advances tournament rounds |

---

## 12. Database Conventions

- **Driver:** `pgx` via `sqlx` (PostgreSQL only).
- **Migrations:** Raw SQL files in `backend/database/migrations/`. Run manually or via scripts.
- **No ORM.** All queries are raw SQL strings embedded in repository functions.
- **Naming:** Table names are `snake_case` plural (e.g. `song_ratings`, `role_user`).
- **Pivot tables:** Use the standard `{entity_a}_{entity_b}` naming (e.g. `badge_user`, `role_user`).
- **UUIDs:** Generated in Go with `github.com/google/uuid` before insertion. Not DB-generated.
- **Timestamps:** `created_at` and `updated_at` are always `CURRENT_TIMESTAMP` in SQL, managed at insert/update time.
- **Search:** GIN indexes for vector search on `song_slugs_combined`.

---

## 13. Testing Strategy

The backend employs a multi-layered testing strategy to ensure correctness and security:

### DTO Security Tests
- **Location:** `internal/dto/mapper_test.go`
- **Purpose:** Ensures no internal numeric IDs are leaked to the API.
- **Tooling:** Uses `testutil.AssertNoInternalIDs` helper.

### Auth Integration Tests
- **Location:** `internal/usecase/auth/auth_usecase_test.go` (package `auth_test`)
- **Purpose:** Validates the full authentication lifecycle:
    - User registration and login.
    - Social account linking (AniList, Google, Discord).
    - OAuth auto-registration.
    - Token encryption/decryption validation.
- **Tooling:** Uses `testutil` OAuth mocks (`MockAnilistClient`, etc.) and `MockUserRepository` hooks for isolated verification.

### Running Tests
Use the following command to run all usecase tests:
```bash
go test ./internal/usecase/... -v
```
