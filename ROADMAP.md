# AniRank — Product Roadmap

> Living document. Updated as features are planned, started, or shipped.
> Status legend: ✅ Done · 🚧 In Progress · 🔲 Planned · 💡 Idea / Exploratory

---

## Current Maturity: Beta Robusta (6.5 / 10)

- Backend architecture: production-grade Clean Architecture with PostgreSQL + Redis + S3.
- Frontend: functional but with UI inconsistency debt being resolved.
- Missing: email infrastructure, automated testing, rate limiting on sensitive routes.

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

### Email Infrastructure

- 🔲 **SMTP integration** (Resend / Postmark / Brevo) for transactional emails
- 🔲 Password reset flow for native (email/password) accounts
- 🔲 Email verification on registration
- 🔲 Notification digests (optional weekly email summary)

### Frontend Polish

- 🚧 Border radius unification (Editorial system: `rounded-sm` standard, see `DESIGN.md`) only public side
- 🔲 Skeleton loading states across all data-fetching views
- ✅ Error boundary pages (404 Not Found, 500 Server Error) with branded design
- 🔲 Toast notification system audit (ensure all async actions provide feedback)
- 🔲 Empty state illustrations for search, playlists, anime list
- 🔲 Responsive audit on tablet breakpoints (iPad landscape)
- 🔲 Accessibility pass (keyboard navigation, ARIA labels, focus rings)

### Notifications

- ✅ **Real-time delivery** via Server-Sent Events (SSE)
- ✅ Notification preferences (opt-out per notification type)

### Search

- ✅ Full-text search upgrade (PostgreSQL `tsvector`)
  - Implemented via unified search index with GIN indexing and relevance ranking (`ts_rank`).
- ✅ Filters on song search (by type, format, season, year, artist)
- ✅ Search history (client-side, localStorage)

### Performance

- 🔲 API response caching strategy (define TTLs per endpoint in Redis)
- ✅ Image lazy loading on all catalog grids and components (Artist/Song/Anime cards)
- 🔲 `srcset` support for dynamic resolutions (requires backend thumbnail generation)
- 🔲 Pagination cursor-based migration for large datasets (currently offset-based)
- 🔲 Database index audit (cover common WHERE/ORDER BY patterns)

### Quality & Testing

- 🔲 **Integration tests** for auth usecase (login, register, OAuth flows)
- 🔲 **Handler tests** for protected routes (auth middleware behavior)
- 🔲 **DTO mapper tests** (ensure no numeric ID leaks in serialization)
- 🔲 **Frontend component tests** (Vitest + Svelte Testing Library)
- 🔲 CI pipeline (GitHub Actions: lint + test on PR)

---

## Phase 3 — Feature Expansion 🔲 (Planned)

### Social & Community

- 🔲 **Profile themes / palette switcher** — per-user CSS accent color (roses, blues, greens)
- 🔲 **Discord account linking** — social identity and future role sync
- 🔲 **Comment reactions** (emoji reactions on individual comments)
- 🔲 **Shared playlists** (collaborative playlist editing)
- 🔲 **User-to-user recommendations** ("You might like this, based on your ratings")

### Content Expansion

- 🔲 **Artist biographies** (rich text, social links, discography view)
- 🔲 **Song lyrics** (optional lyrics display with attribution)
- 🔲 **Alternate titles / romanization** for song names
- 🔲 **Soundtrack / OST type** as a 4th song category (beyond OP/ED/IN)
- 🔲 **Studio and Producer profile pages** (currently listed but no detail view)

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
| 🔴 High   | No SMTP — password reset broken for native accounts                             | `backend/internal/usecase/auth/` |
| 🟡 Medium | No automated tests anywhere in the project                                      | All                              |
| 🟡 Medium | Offset pagination on large tables (songs, anime) will degrade                   | `repository/postgres/`           |
| 🟡 Medium | `ILIKE` search will slow on large datasets                                      | `search_usecase.go`              |
| 🟢 Low    | `product_roadmap.md` in root is now superseded by this file                     | Root                             |
| 🟢 Low    | Several `rounded-xl` / `rounded-full` inconsistencies remain in non-index pages | `frontend/src/`                  |
