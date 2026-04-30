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
| Cache | Redis (with `ResilientStorage` fallback to Memory) |
| Storage | S3-compatible (MinIO in Docker, AWS S3 in production) |
| Auth | JWT HS256 (`golang-jwt/jwt/v5`) |
| Scheduler | `robfig/cron/v3` for background jobs |
| Containerization | Docker Compose (separate `backend` and `frontend` services) |
| OG Images | Custom Go generator (`infrastructure/og`) |
| Resilience | `REDIS_ENABLED` flag to explicitly bypass Redis connectivity |

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

- **Don't** run `migrate reset` on the production database. Only on local/dev environments.

The `HasPermissionMiddleware` checks the authenticated user's permissions (loaded from DB) before allowing access to a specific route.

---

## 8. XP & Gamification System

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

## 9. Error Handling Convention

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

## 10. DTO Layer Rules

- All mappers live in `internal/dto/mapper.go`.
- The `ID` field in any public-facing DTO **must be set to the UUID**, never the numeric ID.
- Token/secret fields (passwords, OAuth tokens) must use `json:"-"` in the domain struct and must NOT appear in any DTO.
- The canonical pattern for mapping a user is: `ToUserMinimalDTO` ⊂ `ToUserDTO` ⊂ `ToAuthUserDTO`.
- **Security Validation:** All mappers MUST be covered by unit tests in `mapper_test.go` using the `testutil.AssertNoInternalIDs` helper. This helper uses reflection to ensure no sequential IDs are leaked through the DTO hierarchy.

---

## 11. External Integrations

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

## 12. Background Jobs

Managed by `robfig/cron/v3`. Defined in `internal/jobs/`.

| Job | Schedule | Purpose |
|-----|----------|---------|
| `TakeRankingSnapshot` | Configurable | Snapshots current song rating positions |
| `AdvanceTournament` | Configurable | Auto-advances tournament rounds |

---

## 13. Database Conventions

- **Driver:** `pgx` via `sqlx` (PostgreSQL only).
- **Migrations:** Raw SQL files in `backend/database/migrations/`. Run manually or via scripts.
- **No ORM.** All queries are raw SQL strings embedded in repository functions.
- **Naming:** Table names are `snake_case` plural (e.g. `song_ratings`, `role_user`).
- **Pivot tables:** Use the standard `{entity_a}_{entity_b}` naming (e.g. `badge_user`, `role_user`).
- **UUIDs:** Generated in Go with `github.com/google/uuid` before insertion. Not DB-generated.
- **Timestamps:** `created_at` and `updated_at` are always `CURRENT_TIMESTAMP` in SQL, managed at insert/update time.
- **Search:** GIN indexes for vector search on `song_slugs_combined`.

---

## 14. Testing Strategy

The backend employs a multi-layered testing strategy to ensure correctness and security:

### Running Tests
Use the following command to run all usecase tests:
```bash
go test ./internal/usecase/... -v
```

---

## 15. Frontend Testing Strategy

The frontend uses **Vitest** for unit and component testing, ensuring that Svelte 5 reactive states (Runes) and UI components behave as expected.

### Core Stack
- **Runner:** Vitest
- **DOM Simulation:** jsdom
- **Mocking:** 
  - **API:** MSW (Mock Service Worker) for intercepting Axios calls.
  - **Environment:** Manual mocks for `$app/state`, `$app/stores`, and `localStorage` located in `src/tests/setup.ts`.

### Svelte 5 Setup
Due to Svelte 5's architecture, Vitest is configured with `resolve.conditions: ['browser', 'development']` to ensure that client-side lifecycle functions like `onMount` are correctly loaded instead of the SSR versions.

### Testing Locations
- **State Logic:** `src/lib/state/*.test.ts` (e.g., `auth.test.ts` for `$state` and `$derived` logic).
- **Component Behavior:** `src/tests/components/*.test.ts` (using Svelte Testing Library to interact with the DOM).

### Optimistic UI Validation
- Tests for interactions (likes, favorites) must verify that the UI updates immediately (optimistically) before MSW resolves the network request, and that it correctly rolls back upon API failure.

---

## 16. Performance Best Practices

### Data Fetching
- **No N+1 Queries:** Use Cases must implement bulk hydration for lists.
- **Repository Joins:** Repositories should include basic relations (e.g., `Song` should JOIN `Anime` in `GetMany`) to avoid secondary lookups.
- **Cache Policy:** Use short-lived (5-30m) Redis/Memory cache for heavy entry points like `/api/home` or `/api/ranking`.
- **Resilient Cache:** All Redis operations fallback to local memory if the service is unreachable (see `cache.NewResilientStorage`).

### Infrastructure Health
- **Timeouts:** All external I/O (Redis, S3, AniList) must have a context timeout (typically 1s-3s).
- **Graceful Degradation:** The system must remain functional (even if slower) if non-critical services like Redis or MinIO are down.
