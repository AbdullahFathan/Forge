# WorkSpace frontend

Vite + React + TypeScript UI for WorkSpace. Product: [prd.md](../../docs/prd.md). Visual rules: [desain_system.md](../../docs/desain_system.md). Tickets: [milestone-fe.md](../../docs/milestone-fe.md). API: [backend.md](../../docs/backend.md).

## Run

```bash
cd forge/forge-fe
cp .env.example .env
bun install
bun run dev
```

App: http://localhost:3000

`VITE_API_URL` points at the Go API (default `http://localhost:8080`). The API `CORS_ORIGINS` must include the Vite origin (`http://localhost:3000`) because the refresh token is an HttpOnly cookie and the client sends `withCredentials`. The access token stays in memory and is never written to `localStorage`.

Bootstrap admin (from the API seed): `admin@workspace.local` / `ChangeMeAdmin1`.

```bash
bun run lint
bun run build
bun test
```

`bun test` runs Vitest (login schema, permission checks, response unwrap, and a single 401 refresh retry).

## Folder tree

Paths under `src/`:

- `app/` — `main.tsx`, `App.tsx`, `router.tsx`, providers, layouts
- `features/<name>/{api,components,hooks,pages}` — auth, dashboard, profile, users, departments, and later domains
- `components/ui` — shadcn primitives
- `components/common` — shell, tables, badges
- `services/api` — Axios client, interceptors, endpoints
- `services/query` — TanStack Query client and keys
- `hooks`, `lib`, `types`, `styles`, `assets`

Do not add a second `src/` shape. `docs/react-struc.txt` is an example only.

A `/dev` route in development shows status and priority badges. Remove it before Phase 5.

## Auth

Login calls `POST /auth/login`. Refresh and logout use the `refresh_token` cookie (`POST /auth/refresh`, `POST /auth/logout`). Navigation permissions come from the access token claims (`perms`), because `GET /users/me` returns the profile without a permission list.
