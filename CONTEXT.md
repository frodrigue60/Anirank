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
- AniList and Google OAuth account linking and unlinking.
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

## 5. Security Headers & Hardening

To achieve a 100% Best Practices score, both SvelteKit and Go enforce identical security policies. **Any change to one must be reflected in the other.**

### Enforcement Locations
- **Frontend:** `frontend/src/hooks.server.ts`
- **Backend:** `backend/internal/delivery/http/middleware/security.go`

### Mandatory Directives
| Header | Value / Standard |
|--------|-----------------|
| `Content-Security-Policy` | `script-src 'self' 'nonce-...'`; `img-src 'self' data: s3-url...` |
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains; preload` |
| `X-Frame-Options` | `DENY` (Prevents clickjacking) |
| `Cross-Origin-Opener-Policy`| `same-origin` |
| `X-Content-Type-Options` | `nosniff` |

### CSP Local Development
For local development involving media assets, the `img-src` directive MUST include `http://localhost:9000` to allow connections to the local MinIO container.

---

## 6. Domain Entities Summary

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
| `Partner` | `id`, `uuid`, `name`, `url`, `banner` | External community/source links. |

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
- Tests for interactions (likes, favorites, ratings) must verify that the UI updates immediately (optimistically) before MSW resolves the network request, and that it correctly rolls back upon API failure.

### Web Animations API Caveat
- Svelte 5 transitions (e.g., `fade`, `scale`) rely on the Web Animations API. Since JSDOM does not support it, `Element.prototype.animate` must be mocked in `src/tests/setup.ts` to prevent test failures.

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
---

## 17. Automated Moderation (Automod)

The system features a centralized moderation engine (`ModerationUsecase`) that validates user interactions (ratings, comments) based on reputation and behavioral rules.

### Reputation System (TruthScore)
Every user has a `truth_score` (default: 100).
- **Incentives:** Successful reports (+5), positive engagement.
- **Penalties:** Rejected reports (-5), accepted reports against the user (-10), spam detection.
- **Auto-Recovery:** TruthScore is dynamic and can be recovered through positive contributions.

### Core Moderation Rules
| Rule | Condition | Action |
|------|-----------|--------|
| **Softban** | `TruthScore < 30` AND `PendingReports > 3` | Block all new interactions (Ratings, Comments). |
| **Rate Limit** | `Level < 5` (120s), `Level < 10` (60s) | Prevents rapid spamming. |
| **Link Filter** | `Level < 5` | Rejects content containing URLs. |
| **Auto-Shadowban** | `Level 5-10` + URL | Content is allowed but marked `is_shadowbanned=true` (hidden from others). |
| **Truth Shadowban** | `TruthScore < 50` | Automatically shadowbans all subsequent user interactions. |

### Implementation Details
- **Validation:** Controlled by `ValidateInteraction(ctx, userID, content)`.
- **Enforcement:** Applied in `CreateComment` and `RateSong` before data persistence.
- **Auto-Lifting:** Softbans are automatically lifted if `TruthScore >= 40`.
- **Shadowban Logic:** Shadowbanned comments/ratings are only visible to the author and staff. They do not affect public ranking averages or comment counts.
### Immutable Moderation Snapshots
To ensure a tamper-proof audit trail for administrative review, the system captures **Immutable Snapshots** of reported entities at the exact moment a report is created.
- **Mechanism:** The `ModerationUsecase` fetches the current state of the reported entity (Song, Comment, or User) from the database, serializes it into a JSON string, and stores it in the `snapshot` column of the report.
- **Integrity:** Once a report is created, its snapshot remains unchanged even if the original entity is edited or deleted.
- **Workflow:** `Fetch -> Marshal -> Assign`. This allows staff to see the exact content that was reported, regardless of subsequent changes by the user.
---

## 18. Database Migrations & Idempotency

AniRank uses a custom raw-SQL migration system. All migrations must be **idempotent**, meaning they can be executed multiple times without causing errors or data corruption.

### Migration Best Practices

1.  **File Naming:** Use the format `YYYYMMDDHHMM_description.sql` (e.g., `202405141100_add_shadowban_to_reactions.sql`).
2.  **Idempotent Columns:** Use `ALTER TABLE ... ADD COLUMN IF NOT EXISTS ...`.
3.  **Idempotent Indices:** Use `CREATE INDEX IF NOT EXISTS ...`.
4.  **Complex Logic:** Wrap non-native idempotent operations in a `DO $$` block:
    ```sql
    DO $$ 
    BEGIN
        IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='my_table' AND column_name='my_col') THEN
            ALTER TABLE my_table ADD COLUMN my_col BOOLEAN DEFAULT false;
        END IF;
    END $$;
    ```
5.  **No Down Migrations:** The system currently only supports forward migrations. To revert a change, create a new "fix" migration.
6.  **Verification:** Always test migrations locally using `go run scripts/run_migrations.go` before committing.

### Handling Schema Drift
If a migration was manually recorded as "run" in the `migrations` table but failed to apply all changes (e.g., due to manual DB edits), create a **reconciliation migration** that uses the `IF NOT EXISTS` pattern to ensure all expected columns are present.

---

## 19. AMQ & WebSocket Architecture

The **Anime Music Quiz (AMQ)** feature is a real-time multiplayer game built on a pure in-memory actor model over WebSockets. There is **no database persistence** for active rooms or sessions — all state lives in RAM and is lost on server restart.

### Key Files

| File | Responsibility |
|------|---------------|
| `backend/internal/usecase/amq/lobby_manager.go` | Global registry of all active rooms. Creates, finds, and destroys rooms. Runs the room cleanup loop. |
| `backend/internal/usecase/amq/lobby_room.go` | The room actor: holds all game state, runs the single-goroutine event loop, and handles all game transitions. |
| `backend/internal/delivery/http/v1/amq_handler.go` | HTTP → WS upgrade handler. Reads raw WS messages, translates them to `RoomEvent`s, and dispatches via `SendEvent`. |
| `frontend/src/routes/(app)/amq/+page.svelte` | Lobby browser: lists public rooms, room creation form. |
| `frontend/src/routes/(app)/amq/[roomId]/+page.svelte` | Game room page: manages WS connection lifecycle, renders game state. |

---

### Backend: Actor Model

Each `LobbyRoom` is an isolated actor. It owns:
- `Players map[string]*domain.AMQPlayer` — session-keyed player state (includes host, ready, score, offline flags).
- `Conns map[string]WSConn` — session-keyed WebSocket connections.
- `EventChan chan RoomEvent` — buffered channel (capacity 100) for serialized event processing.
- `mu sync.RWMutex` — protects all room state.

**The cardinal rule: all mutations happen through the event loop.**

```
HTTP Handler → room.SendEvent(ev) → EventChan → run() goroutine → handleEvent(ev) → mutate state → broadcast()
```

#### Event Loop (`run()`)

```go
func (r *LobbyRoom) run() {
    cleanupTicker := time.NewTicker(10 * time.Second)
    for {
        stopped := false
        func() {
            defer recover() // Panics are logged, not fatal
            select {
            case ev, ok := <-r.EventChan:
                if !ok { stopped = true; return }
                r.handleEvent(ev)
                r.mu.Lock(); r.LastActive = time.Now(); r.mu.Unlock()
            case <-cleanupTicker.C:
                r.cleanupOfflinePlayers()
            }
        }()
        if stopped { return }
    }
}
```

- **Single goroutine**: events are processed sequentially — no race conditions within event handlers.
- **Panic-safe**: `recover()` prevents any single event from crashing the loop.
- **Cleanup ticker**: every 10s, purges players offline for >60s. If all players are purged, `ShouldDestroy()` returns true and `LobbyManager` destroys the room every 30s.

#### Event Types

| Constant | Trigger | Handler |
|----------|---------|---------|
| `EvJoin` | WS connection established | `handleJoin` — reassociates or creates player |
| `EvLeave` | WS connection closed | `handleLeave` — marks player offline |
| `EvReady` | `player_ready_toggle` | `handleReady` — toggles ready state |
| `EvConfigUpdate` | `update_lobby_config` (host only) | `handleConfigUpdate` |
| `EvStartGame` | `start_game` (host only) | `handleStartGame` → goroutine builds song pool → `EvPoolLoaded` |
| `EvPoolLoaded` | Internal (pool ready) | `startRound` |
| `EvSubmitGuess` | `submit_guess` | `handleSubmitGuess` |
| `EvTimerExpired` | Internal timer | `handleTimerExpired` |
| `EvSkipSummary` | `skip_summary` | `handleSkipSummary` |
| `EvResetToLobby` | `reset_to_lobby` | `handleResetToLobby` / `forceResetToLobby` |
| `EvChat` | `send_chat_message` | `handleChat` |
| `EvTransferHost` | `transfer_host` (host only) | `handleTransferHost` |
| `EvSelectCandidate` | `select_candidate` | `handleSelectCandidate` (save mode) |
| `EvSkipWinnerPlayback` | `skip_winner_playback` (host) | `handleSkipSavePlayback` (save mode) |

---

### Locking Rules (Critical)

The `LobbyRoom` has two separate mutexes:
- `r.mu sync.RWMutex` — protects room state (`Players`, `Conns`, `Config`, `Status`, etc.).
- `LobbyManager.mu sync.RWMutex` — protects the rooms map.

**The invariant that must never be broken:**

```
Any function that holds r.mu.Lock() MUST release it before calling broadcast() or sendTo().
```

**Why:** `broadcast()` and `sendTo()` intentionally use `r.mu.RLock()` (not `Lock()`), but a goroutine cannot acquire an `RLock` on an `RWMutex` it already holds as a `Lock`. This would deadlock. Additionally, using `Lock()` in `broadcast()` would conflict with `LobbyManager.cleanupRooms()`, which holds `LobbyManager.mu.Lock()` while calling `room.ShouldDestroy()` → `r.mu.RLock()`. If `broadcast()` tried `r.mu.Lock()` simultaneously, it would form a partial lock ordering that risks deadlock under load.

**Correct pattern (used in every handler):**
```go
func (r *LobbyRoom) handleSomething(sessionID string) {
    r.mu.Lock()
    // ... mutate state ...
    r.mu.Unlock()           // ← ALWAYS unlock before broadcast
    r.broadcast("lobby_state_update", r.getRoomStatePayload())
}
```

---

### WebSocket Connection Lifecycle

```
1. Client calls POST /api/amq/rooms              → creates room, gets room_id
2. Client calls GET  /api/amq/ws/{roomId}?...    → HTTP → WS upgrade (WSUpgrade middleware)
3. WSHandler runs:
   a. Validates JWT token (optional auth) → resolves *domain.User or nil
   b. Generates sessionID = uuid.New()
   c. Wraps *websocket.Conn in wsConnWrapper (implements WSConn interface)
   d. Calls LobbyManager.JoinRoom() → dispatches EvJoin
   e. defer LobbyManager.LeaveRoom() → dispatches EvLeave on disconnect
   f. Reader loop: reads JSON messages → dispatches corresponding events
```

**Session Identity:**
- Each WS connection gets a new `sessionID` (UUID) even if it's the same user reconnecting.
- `handleJoin` resolves the actual player by:
  - **Authenticated users**: matched by `UserUUID` across all sessions.
  - **Guests**: matched by `DeviceID` (stored in `localStorage` as `amq_device_id`).
- On match: old session is deleted, player struct is moved to new sessionID → reconnection is seamless.
- On no match: new player is created and added to the room.

**Host Persistence:**
- `ensureHostActive()` is called after every `EvJoin` and `EvLeave`.
- If the current host goes offline, the first online non-spectator player is promoted automatically.
- `handleTransferHost` allows the current host to manually promote another online player.

---

### WebSocket Message Protocol

All messages use the envelope format:
```json
{ "type": "<event_type>", "payload": { ... } }
```

#### Server → Client messages

| Type | Trigger | Payload |
|------|---------|---------|
| `lobby_state_update` | Any state change (join, leave, ready, config, transfer) | Full room snapshot (see below) |
| `round_start` | New round begins | `{ audio_url, current_round, max_rounds, guess_time, start_percent, options[] }` |
| `round_ended` | Timer expires or all locked | `{ song: SongDTO, player_results[], correct_slug }` |
| `chat_message` | Chat or system event | `{ sender, text, type: "user"|"system", timestamp }` |
| `error` | Fatal room error | `string` error message |

#### `lobby_state_update` payload shape
```json
{
  "room_id":    "FA2705CC",
  "status":     "lobby | playing | reveal | finished",
  "config":     { "name", "max_rounds", "guess_time", "reveal_time", "theme_type", "game_type", "personalized_pool", "private" },
  "players":    [ { "session_id", "nickname", "user_uuid", "device_id", "is_host", "is_ready", "is_spectator", "offline", "score", "locked" } ],
  "spectators": [ ... same shape as players ... ],
  "timer_left": 15,
  "round_data": null
}
```

#### Client → Server messages

| Type | Payload | Permission |
|------|---------|-----------|
| `player_ready_toggle` | (none) | Any player |
| `submit_guess` | `{ "anime_slug": string }` | Any player |
| `update_lobby_config` | `domain.AMQConfig` | Host only |
| `start_game` | (none) | Host only |
| `skip_summary` | (none) | Any player (vote to skip) |
| `reset_to_lobby` | (none) | Host (full reset) or any player (self-ready reset) |
| `transfer_host` | `{ "target_session_id": string }` | Host only |
| `send_chat_message` | `{ "text": string }` | Any player |

---

### Game State Machine

```
lobby ──[start_game + all ready]──► playing ──[timer expires / all locked]──► reveal
  ▲                                                                              │
  │                              ◄──────[skip_summary / reveal timer]───────────┘
  │
  └──[max_rounds reached]──► finished ──[reset_to_lobby]──► lobby
```

- **`lobby`**: players join, set ready, host configures. Start requires all online non-spectator players to be ready.
- **`playing`**: song is playing, players submit guesses. Timer counts down. Ends when all players lock OR timer hits 0.
- **`reveal`**: correct answer shown, scores updated, XP awarded. Ends when reveal timer expires or majority skips.
- **`finished`**: game over screen. Host (or any player via force) can reset to lobby.

---

### Save Mode (`save-4` / `save-6`)

Social voting mode: no objective correct answer. Each round presents 4 or 6 candidate songs (local video only). Players vote during rotating previews; winner(s) = most votes (ties include 0 votes).

#### Config (save-only fields)

| Field | JSON | Notes |
|-------|------|-------|
| `PreviewSeconds` | `preview_seconds` | Per-candidate preview and winner playback; sanitized **10–15** (default 12) |
| `ThemeDistribution` | `theme_distribution` | `"random"` (default) or `"balanced"` — how theme kinds rotate across rounds |
| `PersonalizedPool` | `personalized_pool` | Always forced **false** for save modes |

Sanitization: `sanitizeSaveConfig()` in `save_pool_builder.go` / `save_round.go`. Max rounds **5–30**.

#### Round pool (`save_pool_builder.go`)

- Theme kinds: `artist`, `year`, `season`, `anime`, `genre`.
- Each round is **all OP or all ED** (`RoundThemeType`); lobby `"both"` picks OP or ED randomly per round.
- Theme keys deduplicated within a game (`ThemeKey`).
- Minimum **2** candidates per round; ideal count is 4 (`save-4`) or 6 (`save-6`). Anime anchors require ≥2 distinct `(type:theme_num)` themes.
- Fallback rounds (`IsFallback: true`) use the global local-video pool only — **no embed URLs**.
- `balanced` distribution prioritizes one kind per round slot before retrying others.

#### Save state machine

```
lobby ──[start_game]──► playing (preview_select)
                              │
                    [preview_step timer × N candidates]
                              │
                         tally votes
                              │
                         winner_playback
                              │
                    [winner_step timer × tied winners]
                              │
                         finishSaveRound ──► next round OR finished
```

Phases stored in `LobbyRoom.RoundPhase`: `preview_select`, `winner_playback`.

#### Additional backend events

| Constant | Trigger | Handler |
|----------|---------|---------|
| `EvSelectCandidate` | `select_candidate` WS | `handleSelectCandidate` — toggle vote during preview |
| `EvSkipWinnerPlayback` | `skip_winner_playback` WS | `handleSkipSavePlayback` — host skips remaining winner playback |

Timer types: `preview_step`, `winner_step` (both use `PreviewSeconds`).

#### Additional server → client messages

| Type | When | Key payload fields |
|------|------|-------------------|
| `round_start` | Save round begins | `round_phase`, `preview_index`, `preview_seconds`, `theme_label`, `round_theme_type`, `is_fallback`, `candidates[]` |
| `phase_change` | Preview index or winner playback step | `round_phase`, `preview_index` or `winner_play_index`, optional `votes`, `winners` |
| `round_results` | After preview tally | `votes`, `winners` |

`lobby_state_update` may include `round_data` (full save snapshot for reconnect) and `save_round_history` when `status === "finished"`.

#### Additional client → server messages

| Type | Payload | Permission |
|------|---------|-----------|
| `select_candidate` | `{ "song_uuid": string }` — empty string deselects | Non-spectator players during `preview_select` |
| `skip_winner_playback` | (none) | Host only during `winner_playback` |

`skip_summary` remains for quiz **reveal** skip; save mode winner skip uses `skip_winner_playback`.

#### Vote rules

- Last `SelectedSongUUID` at preview end counts; offline/spectator votes ignored.
- Ties at max votes → all tied candidates are winners (each gets winner playback).
- Zero total votes → all candidates tie as winners.

#### Key files

| File | Role |
|------|------|
| `save_pool_builder.go` | Thematic pool + fallback |
| `save_round.go` | Preview/tally/playback loop, history, config sanitize |
| `song_repository_amq_save.go` | Anchor queries + local-video-only eligibility |

---

### Frontend Integration (`[roomId]/+page.svelte`)

The room page manages a single WebSocket connection with automatic reconnection.

**Connection lifecycle:**
- Initiated in a `$effect` that waits for `authState.loading === false`.
- Authenticated users connect immediately; guests must set a nickname first (persisted in `localStorage`).
- On unclean close (`!event.wasClean`): auto-reconnects after 3s if `roomState` is not null; redirects to `/amq` if no state yet.
- A `connectionGeneration` counter prevents stale handlers from triggering after reconnect.

**Key reactive state:**
```typescript
let roomState = $state<any>(null);            // Full server snapshot
let players   = $derived(roomState?.players || []);
let selfPlayer = $derived.by(() => {          // Current user's player object
  return players.find(p =>
    authState.isAuthenticated
      ? p.user_uuid === authState.user?.uuid
      : p.device_id === deviceId
  );
});
let playersVersion = $state(0);              // Incremented on each lobby_state_update
```

**Svelte 5 `#each` key pattern:**
```svelte
{#each sortedPlayers as player (player.session_id + '-' + playersVersion)}
```
The compound key forces Svelte to re-diff all player items when `playersVersion` increments, ensuring fields like `is_host` and `is_ready` visually update even when `session_id` is unchanged.

---

### Known Pitfalls & Decisions

| Pitfall | Decision |
|---------|---------|
| `broadcast()` originally used `r.mu.Lock()` | Changed to `r.mu.RLock()` (copy Conns map, then write outside lock) to eliminate deadlock with cleanup goroutine |
| `run()` goroutine dying on panic | Wrapped each iteration in an anonymous func with `defer recover()`. The `stopped` bool propagates a clean channel-close signal out of the closure |
| Svelte `#each` not re-rendering on `is_host` change | Added `playersVersion` state that increments on every `lobby_state_update`, used as part of the compound `#each` key |
| Guest player duplication on reconnect | `handleJoin` first checks for an existing offline player matching `DeviceID`. If found, transfers state to new sessionID instead of creating a new entry |
| Stale host after all-leave / all-rejoin | `ensureHostActive()` runs after every `EvJoin` and `EvLeave`. It promotes the first online non-spectator player if no online host exists |
| Room surviving after everyone leaves | `cleanupOfflinePlayers()` (every 10s in event loop) purges players offline >60s. `LobbyManager.cleanupRooms()` (every 30s, external goroutine) destroys rooms where `ShouldDestroy()` returns true (0 online players AND inactive >2min) |
