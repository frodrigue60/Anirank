# AniRank — Product Roadmap

> Living document. Updated as features are planned, started, or shipped.
> Status legend: ✅ Done · 🚧 In Progress · 🔲 Planned · 💡 Idea / Exploratory

---

## Current Maturity: Beta Robusta (6.5 / 10)

- Backend architecture: production-grade Clean Architecture with PostgreSQL + Redis + S3.
- Frontend: functional with consistent UI and integrated transactional email workflows.
- Missing: notification digests, automated CI/CD pipeline, and public API documentation.

---

## Phase 1 — Core MVP ✅ (Shipped)

### Authentication & Profiles

- ✅ Email/Password registration and login
- ✅ Google OAuth login and account linking
- ✅ AniList OAuth login and account linking
- ✅ JWT-based sessions (UUID-only payload, no sequential ID leakage)
- ✅ RBAC (Roles: owner, admin, editor, creator) with granular permissions
- ✅ User avatar and banner upload (S3/MinIO)
- ✅ Score format preference (POINT_5, POINT_10, POINT_10_DECIMAL, POINT_100)
- ✅ Profile customization (about text, profile accent color)
- ✅ Password reset via email (dependent on Google for Google-linked accounts)

### Content Engine

- ✅ Hierarchical catalog: Anime → Songs → Variants → Videos
- ✅ Song types: OP, ED, IN (Opening, Ending, Insert)
- ✅ Anime formats, genres, seasons, years (fully managed taxonomies)
- ✅ Artist profiles with avatar and song count
- ✅ Global search (anime, songs, artists, users)
- ✅ Slug-based URLs for all public entities
- ✅ Cascading visibility control (admin toggle)
- ✅ AniList anime metadata hydration (batch import from AniList API)
- ✅ AnimeThemes.moe song data hydration

### Interaction & Social

- ✅ Song ratings with score formats
- ✅ Song reactions (emoji-based)
- ✅ Song comments and replies (threaded)
- ✅ Song and artist favorites
- ✅ User follows (follower/following system)
- ✅ In-app notifications (ratings, comments, follows)
- ✅ User leaderboard (by XP, ratings, comments)
- ✅ Activity feed (recent ratings and community actions)

### Gamification (XP System)

- ✅ XP awarded for: rating, commenting, replying, daily login
- ✅ Level progression (1–100, DB-configurable thresholds)
- ✅ Cooldown protection per activity to prevent spam
- ✅ Badges (automatic by requirement_type, manual by admin)
- ✅ Badge display on user profiles

### Discovery & Home

- ✅ Featured song (hero section on index)
- ✅ Weekly ranking (Top OPs / Top EDs)
- ✅ New releases carousel
- ✅ Most popular / Most viewed carousels
- ✅ Featured artists sidebar (algorithmic: favorites + song count)
- ✅ AniList list sync (user's anime list with DB match/missing distinction)
- ✅ Request system for missing content from AniList sync

### Admin Panel

- ✅ Full CRUD: Anime, Songs, Variants, Artists, Taxonomies, Playlists, Badges, Tournaments
- ✅ User management (create, edit, role assignment, password reset)
- ✅ Moderation tickets (song, comment, user reports)
- ✅ User request management (content requests from users)
- ✅ Audit log (admin action tracking)
- ✅ Dashboard stats
- ✅ XP activity configuration
- ✅ Permission editor per role

### Playlists

- ✅ User-created playlists (CRUD)
- ✅ Song addition, removal, reordering
- ✅ Public playlist pages with OG metadata

### Infrastructure

- ✅ Docker Compose (backend + frontend + PostgreSQL + Redis + MinIO)
- ✅ S3-compatible media storage (MinIO dev, AWS S3 prod)
- ✅ Redis cache (optional, NoOpCache fallback)
- ✅ OG image generation for all public entities (anime, song, artist, playlist, user, home)
- ✅ SEO bot proxy (pre-rendered metadata for social crawlers)
- ✅ Sitemap generation (XML + JSON)
- ✅ Background job scheduler (cron: ranking snapshots, tournament auto-advance)
- ✅ CORS configuration

### Tournaments

- ✅ Bracket-style community tournaments
- ✅ Matchup voting
- ✅ Auto-advance via cron job
- ✅ Admin seed and management

---

## Phase 2 — Stability & Polish 🚧 (Active / Up Next)

### Security Hardening

- ✅ UUID-only external API (no sequential ID leakage)
- ✅ Empty UUID guard in AuthMiddleware (prevents 500 from stale tokens)
- ✅ **Rate limiting** on sensitive routes (login, register, OAuth callbacks)
- ✅ **CSRF protection** audit for state-mutating endpoints
- ✅ **Input sanitization** review for comment/about text fields (XSS vectors)

- ✅ **Immutable Moderation Snapshots**: Automatically capture the state of reported entities (songs, comments, users) at the time of reporting to ensure a tamper-proof audit trail.
- 🔲 **Advanced Moderation (AutoMod)**: 🚧 In Progress

- ✅ **Shadow Banning**: Automatically hide content from toxic users without alerting them (Internal `shadow_banned` flag).
- 🔲 **Profanity & Scam Filter**: Regex-based blocking of harmful keywords and known scam URLs.
- ✅ **Reputation Gates (XP-based)**: Restrict link-sharing and frequent posting for users below Nivel 5.
- 🔲 **Auto-Hide Threshold**: Automatically hide comments/users after X reports until admin review.
- 🔲 **Spam Detection**: Velocity limits and duplicate content detection across different entities.
- 🔲 **AI Image Moderation**: Automated NSFW scanning for custom avatars and banners.
- 🔲 **Ban Manager**: Advanced admin UI for temporary/permanent bans, IP blocking, and moderation notes.

- ✅ **SMTP integration** (Resend) for transactional emails
- ✅ Password reset flow for native (email/password) accounts
- ✅ Email verification workflow (User-initiated from settings)
- 🔲 Notification digests (optional weekly email summary)

### Frontend Polish

- ✅ **Border radius unification** (Editorial system: `rounded-sm` standard, see `DESIGN.md`) only public side
- 🔲 Skeleton loading states across all data-fetching views
- ✅ Error boundary pages (404 Not Found, 500 Server Error) with branded design
- ✅ **Toast notification system audit** (feedback on all async actions)
- ✅ **Empty state system** for search, playlists, profiles and activity feed
- ✅ **Responsive audit** on tablet/mobile (table horizontal scroll & modal padding)
- ✅ **Accessibility pass** (ARIA labels, descriptive titles and icon context)

### Notifications

- ✅ **Real-time delivery** via Server-Sent Events (SSE)
- ✅ Notification preferences (opt-out per notification type)

### Search

- ✅ Full-text search upgrade (PostgreSQL `tsvector`)
  - Implemented via unified search index with GIN indexing and relevance ranking (`ts_rank`).
- ✅ Filters on song search (by type, format, season, year, artist)
- ✅ Search history (client-side, localStorage)

### Performance

- ✅ API response caching strategy (Redis-backed with per-endpoint TTLs and bypass logic)
- ✅ Image lazy loading on all catalog grids and components (Artist/Song/Anime cards)
- ✅ `srcset` support for dynamic resolutions (backend thumbnail generation via `MediaService`)
- ✅ **No N+1 Batch Hydration**: Eliminated N+1 query patterns in Activity Feed, Catalog, and Rankings, achieving sub-50ms latencies (down from ~220ms).
- ✅ **Pagination cursor-based migration** for large datasets (Backend implementation complete)
- ✅ **Database Index & Performance Audit**
  - [x] **Large Scale Seeder**: Created Go script to populate DB with ~100k songs and ~10k users.
  - [x] **Query Profiling**: Audited critical paths using `EXPLAIN ANALYZE` (Rankings, Feed).
  - [x] **Composite Indexes**: Implemented `idx_songs_status_score_desc`, `idx_songs_seasonal_perf`, and others.
  - [x] **Slow Query Monitor**: Verified sub-1ms response times on optimized queries.
  - [x] **Session Caching**: Implemented UUID-to-UserID caching in Auth Middleware (2m TTL).
  - [x] **Slim DTO Projection**: Implemented lightweight Song DTOs for list views, reducing payload size by ~40%.
  - [x] **Usecase Pruning**: Disabled heavy variant loading for catalog listings.

- ✅ **Performance & Latency Audit** (Post-Optimization)
  - Verified internal latencies: `/api/home` (21ms), `/api/animes` (32ms), `/api/songs` (69ms).
  - Confirmed session cache hit rate >95% for repeat authenticated requests.
- ✅ **Unit & Integration Testing**
  - All `internal/dto` and `internal/usecase/auth` tests passed.
  - Verified no regressions in core domain logic after DTO slimming and cache implementation.
- ✅ **Integration tests** for auth usecase
  - Validate full registration/login lifecycle including password hashing and JWT generation.
  - Mock OAuth provider responses to verify `anilist`, `google`, and `discord` identity persistence.
  - Test token encryption/decryption for social access tokens.
- ✅ **Handler tests** for protected routes
  - Verified `AuthMiddleware` correctly extracts and validates UUIDs from JWTs.
  - Ensured `OptionalAuthMiddleware` behaves as expected when no token is present.
  - Tested endpoint responses for unauthorized (401) and forbidden (403) access attempts.
- ✅ **DTO mapper tests**
  - **Critical**: Verify that `ToUserDTO` and `ToUserMinimalDTO` never include the numeric `ID` field.
  - **Security**: Added reflection-based helper (`testutil.AssertNoInternalIDs`) to detect any `uint64` leak in public DTOs.
  - Ensure all relationships (Badges, Roles) are correctly mapped to their public DTO equivalents.
  - Validate that `UserSocialIdentity` objects are correctly hydrated for the API.
- ✅ **Frontend component tests** (Vitest + Svelte Testing Library)
  - **Auth & Identity Persistence**:
    - Validated `authState` (Svelte 5 Runes) updates globally after successful login.
    - Tested header reactivity (Login/Register buttons vs. User Profile dropdown).
  - **Account Settings & Social Linkage**:
    - Verified conditional rendering in `AccountSettings` (Link vs. Synced states).
    - Mocked social identity removal and asserted UI clears cached session data.
  - **Admin & CMS Form Validation**:
    - Unit tested complex forms (Anime/Song creation) for field validation and error messaging.
    - Verified interactive elements like "Tags input" and "Artist search" within the admin panel.
  - **Optimistic UI & Rankings**:
    - Asserted that song ratings and favorite toggles reflect state changes instantly before API confirmation.
- 🔲 CI pipeline (GitHub Actions: lint + test on PR)

---

## Phase 3 — Feature Expansion 🔲 (Planned)

### Social & Community

- ✅ **Profile themes / palette switcher** — per-user CSS accent color (roses, blues, greens)

- ✅ **Discord account linking** — social identity and future role sync
- 🔲 **Comment reactions** (emoji reactions on individual comments)
- 🔲 **Shared playlists** (collaborative playlist editing)
- 🔲 **User-to-user recommendations** ("You might like this, based on your ratings")
- 💡 **Activity Notifications** — emails for mentions, replies, and new followers (opt-in)

### Content Expansion

- 🔲 **Artist biographies** (rich text, social links, discography view)
- 🔲 **Song lyrics** (optional lyrics display with attribution)
- ✅ **Alternate titles / romanization** for song names
- ✅ **Soundtrack / OST type** as a 4th song category (beyond OP/ED/IN)
- 🔲 **Studio and Producer profile pages** (currently listed but no detail view)
- 💡 **Welcome Email** — automatic onboarding email after successful verification

### Analytics

- 🔲 **Personal stats dashboard** — user's own rating distribution, genre preferences, activity heatmap
- 🔲 **"Taste profile"** — generated summary of a user's anime music preferences based on ratings

### Discovery

- 🔲 **Recommendation engine** — "Songs you might rate highly" based on collaborative filtering
- 🔲 **"Seasonal wrap-up"** — yearly summary (like Spotify Wrapped) of each user's ratings
- 🔲 **Mood-based playlists** (curated by editorial team or algorithm)

---

## Phase 4 — Platform Maturity 💡 (Exploratory)

### Public API

- 💡 **REST API v1 (public)** — documented OpenAPI spec for third-party integrations
- 💡 **API rate limiting per token** for public API consumers
- 💡 **Webhook support** for community rankings
- 💡 **Mail Deliverability Webhooks** — automated handling of bounces/complaints via Resend

### Mobile

- 💡 **PWA manifest** + service worker for installable mobile experience
- 💡 **Native app** (React Native or Capacitor) — if user base justifies it

### Monetization / Sustainability

- 💡 **Supporter tier** — cosmetic perks (profile frames, exclusive badges, early access)
- 💡 **Patreon/Ko-fi integration** for supporters
- 💡 **Affiliate links** to legal streaming services when displaying anime

### Community Tools

- 💡 **Clubs/Groups** — dedicated spaces for genre or artist fan groups
- 💡 **Community polls** (separate from tournaments, for quick votes)
- 💡 **Editorial blog** — staff-written posts about seasonal rankings, hidden gems

---

## Technical Debt Log

> Known issues to address before v1.0 launch.

| Priority  | Issue                                                                           | Location                         |
| --------- | ------------------------------------------------------------------------------- | -------------------------------- |
| ✅ Done   | Rate limiting implemented for Auth and Public API routes                        | `middleware/rate_limiter.go`     |
| ✅ Done   | SMTP integration — Resend service for reset and verification             | `internal/infrastructure/mail/`  |
| ✅ Done   | Automated testing — integration/unit coverage added for Auth, DTOs, and Repos    | `internal/`                      |
| ✅ Done   | Performance pagination — Cursor-based implementation for large datasets          | `repository/postgres/`           |
| ✅ Done   | Optimized search — GIN Trigram indexes for fast ILIKE results                  | `search_usecase.go`              |
| ✅ Done   | `product_roadmap.md` in root is now superseded by this file                     | Root                             |
| ✅ Done   | Several `rounded-xl` / `rounded-full` inconsistencies remain in non-index pages | `frontend/src/`                  |
