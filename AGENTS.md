# AGENTS.md — AniRank AI Agent Configuration

> This file configures any LLM agent working inside this repository. Read it in full BEFORE writing any code or suggesting changes.

---

## 1. Agent Roles

### `frontend-agent`
Responsible for all SvelteKit UI work inside `frontend/`.
- Must always consult `DESIGN.md` before creating or modifying any component.
- Must use Svelte 5 runes (`$state`, `$derived`, `$props`, `$effect`). Never use Svelte 4 store syntax in new files.
- Must follow the TailwindCSS v4 utility convention already in use.

### `backend-agent`
Responsible for all Go API work inside `backend/`.
- Must always consult `CONTEXT.md` before creating or modifying any domain, usecase, or repository.
- Must follow the Clean Architecture layer rules described in `CONTEXT.md`.
- Must never expose internal sequential numeric IDs to the API layer. See **Security Rules** below.

### `fullstack-agent`
Handles cross-cutting concerns (auth flows, API contract changes, Docker, env files).
- Must consult both `DESIGN.md` and `CONTEXT.md` before making changes.
- Any API contract change (new/removed JSON field) must be reflected in both the backend DTO and the frontend `api.ts` consumer simultaneously.

---

## 2. Mandatory Behaviours (All Agents)

### After Writing Code (Mandatory Verification)
1. **Compilation**: Always run `go build ./cmd/api/main.go` (backend) after any logic change to ensure no type errors.
2. **Backend Tests**: Run `go test ./internal/dto/...` and `go test ./internal/repository/postgres/...` (or relevant package) to verify core logic.
3. **Frontend Tests**: Run `bun run test:unit` if any reactive state or shared component was modified.
4. **Moderation Tests**: Run `go test ./internal/usecase/moderation/...` after any change to moderation rules or `TruthScore` logic.

### Before Writing Code
1. Check if a relevant **KI (Knowledge Item)** exists for the task.
2. Read `DESIGN.md` if the task touches UI.
3. Read `CONTEXT.md` if the task touches backend logic, DB, architecture, or frontend testing.
4. Identify which layer the change belongs to (domain, usecase, repository, delivery, dto) and stay within it.
5. - **Frontend Only:** For critical logic (Auth, Optimistic UI, Admin CRUD), verify if a test file exists in `src/tests/` or `src/lib/state/`.
- **Security Audit:** Any change to security headers in `hooks.server.ts` MUST be mirrored in `backend/internal/delivery/http/middleware/security.go`.
- **Accessibility Audit:** New UI elements must be verified for WCAG 4.5:1 contrast. Use opacities of `70%` or higher for secondary text on dark backgrounds.

### Mandatory Testing (Frontend)
- After modifying a reactive state in `$lib/state/`, you MUST run `bun run test:unit` to ensure no regressions.
- New complex components or interactions MUST include a corresponding `.test.ts` file.
- Use **MSW** for all API mocks in component tests. Never make real API calls during Vitest execution.
- **JSDOM Transition Caveat**: JSDOM does not support the Web Animations API. If a component uses Svelte 5 transitions, ensure `Element.prototype.animate` is mocked in `setup.ts`.
- Ensure that **Optimistic UI** changes are tested for both "Success Path" and "Rollback on Error" scenarios.

### Commit Format
All commits must follow Conventional Commits:
```
<type>(<scope>): <short description>

Types: feat, fix, style, refactor, chore, docs, test
Scope: frontend | backend | auth | db | design | api
```

Examples:
- `feat(auth): add AniList OAuth link flow`
- `fix(backend): patch undefined userRepo in optional middleware`
- `style(frontend): unify border radii to rounded-sm across index page`
- `docs(root): add CONTEXT.md and AGENTS.md agent configuration`

### Docker Build
- After any backend `.go` file change, remind the user to run: `docker-compose up --build -d backend`
- After any frontend file change, the Vite dev server auto-reloads. No rebuild needed during development.
- Production builds use: `docker-compose up --build -d backend frontend`

---

## 3. Security Rules (Critical)

> These rules exist to prevent enumeration attacks on the database. Violating them breaks the security model.

### ❌ NEVER expose sequential numeric IDs in API responses
The internal `uint64` IDs in the database (e.g. `user.ID`, `song.ID`) are **strictly internal**. They must never appear in JSON API responses.

### ✅ ALWAYS use UUIDs for external references
- Every entity that needs public referencing has a `UUID string` field.
- Map `ID → UUID` in the DTO layer (see `backend/internal/dto/mapper.go`).
- Slugs are acceptable for human-readable URLs (e.g. `/users/luis-rodz`).
- **Required Verification:** 
    - Always run `go test ./internal/dto/...` after modifying mappers/DTOs to ensure the security reflection helper passes with zero violations.
    - Always run `go test ./internal/usecase/auth/...` after modifying any authentication or identity logic to ensure security flows (encryption, linking) are intact.
    - Always run `go build ./cmd/api/main.go` and `go vet ./...` after any structural change (new interface method, new file, dependency injection). A clean build is the minimum bar.

### JWT Token Policy
- The JWT payload contains only `user_uuid` (string) and `roles` ([]string). Never put `user_id` (numeric) in the token.
- The `AuthMiddleware` validates the UUID from the token, looks up the user in DB, and injects **both** `user_id` (numeric, for internal repo use) and `user_uuid` (string, for responses) into `c.Locals`.
- `OptionalAuthMiddleware` follows the same pattern but silently skips if no valid token or UUID is present.

### Token Guard: Missing UUID = 401, not 500
- If `claims.UserUUID == ""` (stale pre-refactor token), return `401`, never pass an empty string to the DB.

---

## 4. Frontend Coding Style

| Rule | Value |
|------|-------|
| Framework | SvelteKit (Svelte 5) |
| Package Manager | Bun |
| CSS | TailwindCSS v4 via `app.css` |
| State | Svelte 5 Runes only (`$state`, `$derived`) |
| API calls | Axios wrapper at `$lib/api.ts` |
| Auth state | `$lib/state/auth.svelte.ts` (`authState` singleton) |
| Icons | Material Symbols Outlined (font-based) + Lucide Svelte (component-based) |
| Design token reference | `DESIGN.md` |

**Component File Conventions:**
- Files in `src/lib/components/` are shared, generic components.
- Files in `src/routes/(app)/` are page-level components.
- Admin routes live in `src/routes/(admin)/`.

**Optimistic UI Standard:**
- Call state update/callback *immediately* before the API call.
- Wrap API call in `try/catch`.
- In `catch`, restore the previous state and show a toast error message.

---

## 5. Backend Coding Style

| Rule | Value |
|------|-------|
| Language | Go 1.25+ |
| HTTP Framework | GoFiber v2 |
| DB Driver | `sqlx` + `pgx` (PostgreSQL) |
| ORM | None. Raw SQL only via `sqlx`. |
| Error Pattern | Always return `domain.NewAppError(code, message, err)` from handlers |
| Config | `godotenv` — loads `./backend/.env` |

**Naming Conventions:**
- Repository files: `{entity}_repository.go`
- Usecase files: `{entity}_usecase.go` or grouped in `usecase/{domain}/`
- Handler files: `{entity}_handler.go` inside `delivery/http/v1/`
- DTO files: `{entity}_dto.go` — data transfer objects for the API layer

---

## 6. What NOT to do

- **Don't** use `fmt.Sprintf("%d", id)` to stringify a numeric ID and return it in a JSON response.
- **Don't** add new routes without also adding the corresponding middleware (Auth or OptionalAuth where needed).
- **Don't** use `rounded-full` on any UI element except user/artist avatars and spinner indicators.
- **Don't** use glassmorphism, `backdrop-blur`, or `bg-opacity` in UI components. Use solid surface token colors.
- **Don't** import `strconv` in `mapper.go` — there is no need to convert IDs to strings in the DTO layer.
- **Don't** run `migrate reset` on the production database. Only on local/dev environments.
- **Don't** implement custom social unlinking logic in handlers; use the standard `DELETE /api/auth/:provider/unlink` pattern.
- **Don't** use `{@html}` directly with raw strings. Always wrap the content with `createTrustedHTML()` from `$lib/trusted` to comply with the CSP Trusted Types policy.
- **Don't** read `os.Getenv(...)` inside a UseCase to construct URLs or config values. Inject dependencies (e.g. `MediaService`) via the constructor instead.
- **Don't** add a `case` to the switch in `badge_usecase.go` for new badge types. Implement `BadgeEvaluator` and register it in `buildEvaluators()` in `badge_evaluator.go`.
- **Don't** call `GetBadgesByUserID` (singular) inside a loop over comments or any list. Use `GetBadgesByUserIDs` (plural) for batch fetching.
- **Don't** bypass `moderationUsecase.ValidateInteraction` when creating new types of user-generated content. All public interactions must be filtered.
- **Don't** manually update `user.is_softbanned` in a repository. Use the `checkAndApplyShadowban` (internal) or `UpdateSoftbanStatus` logic to ensure consistency with `TruthScore`.
- **Don't** create database migrations that are not idempotent (always use `IF NOT EXISTS` or `DO $$` blocks).
- **Don't** modify existing migration files that have already been committed; always create a new migration for changes or fixes.

---

## 7. Performance & Resilience Patterns (Mandatory)

### Resilient Storage
- All Redis-backed storage (Rate limiters, App Cache) MUST be wrapped in `cache.NewResilientStorage(primary)`.
- This ensures the application falls back to In-Memory storage if Redis becomes unresponsive.
- Use `REDIS_ENABLED=false` in `.env` to skip Redis connectivity entirely in restricted environments.

### Eager Loading (N+1 Prevention)
- In any UseCase returning a list of entities (Feeds, Rankings, Catalogs, Comments), DO NOT loop and call repositories inside the loop.
- Use the **Batch Enrichment Pattern**:
    1. Collect all required IDs into maps/sets (including nested entities like comment replies).
    2. Fetch all related entities (Users, Songs, Artists, Badges) in bulk using `GetMany` or `GetXxxByIDs`.
    3. Hydrate the original list in a single pass using a separate `hydrateXxx()` helper.
- Target latency for list endpoints: **< 50ms**.
- **Badge enrichment**: Use `UserRepository.GetBadgesByUserIDs(ctx, []uint64)` to batch-load badges for all comment authors at once. Never call `GetBadgesByUserID` (singular) inside a loop.

### Media URL Resolution
- **NEVER** call `os.Getenv("S3_PUBLIC_URL")` or construct media URLs manually inside a UseCase.
- **ALWAYS** use `mediaService.Resolve(*string) *string` or `mediaService.GetURL(string) string` — these are already injected as a dependency and handle both absolute URLs and relative R2/S3 paths transparently.
- The `MediaService` interface is defined in `internal/infrastructure/media_service.go`.

### Badge Evaluator Strategy
- Automatic badge logic lives in `internal/usecase/admin/badge_evaluator.go`.
- To add a new badge trigger type, add a new `BadgeEvaluator` implementation and register it in `buildEvaluators()`. **Do not** add a new `case` to a switch in `badge_usecase.go`.
- The `triggerType` string (e.g. `"ratings"`, `"level"`) must match the `requirement_type` column in the `badges` table.

### Database Migrations
- **ALWAYS** use idempotent SQL syntax.
- **Columns:** `ALTER TABLE ... ADD COLUMN IF NOT EXISTS ...`
- **Indices:** `CREATE INDEX IF NOT EXISTS ...`
- **Complex logic:** Wrap in `DO $$ BEGIN ... END $$;` blocks to check `information_schema` before executing structural changes.
- **Documentation:** Consult Section 18 of `CONTEXT.md` for specific patterns.
