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

### Before Writing Code
1. Check if a relevant **KI (Knowledge Item)** exists for the task.
2. Read `DESIGN.md` if the task touches UI.
3. Read `CONTEXT.md` if the task touches backend logic, DB, architecture, or frontend testing.
4. Identify which layer the change belongs to (domain, usecase, repository, delivery, dto) and stay within it.
5. **Frontend Only:** For critical logic (Auth, Optimistic UI, Admin CRUD), verify if a test file exists in `src/tests/` or `src/lib/state/`.

### Mandatory Testing (Frontend)
- After modifying a reactive state in `$lib/state/`, you MUST run `bun run test:unit` to ensure no regressions.
- New complex components or interactions MUST include a corresponding `.test.ts` file.
- Use **MSW** for all API mocks in component tests. Never make real API calls during Vitest execution.
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
