# WorkSpace backend (Phase 1)

Go API: chi, GORM, Zap, JWT + RBAC. Lives at `forge/forge-be` in the monorepo.

## Run

From this directory (`forge/forge-be`). Compose only starts Postgres, Redis, and RustFS. The API runs on the host with `go run` during development.

```bash
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

Then:

- Health: `GET http://localhost:8080/health`
- Swagger: http://localhost:8080/swagger/index.html
- Login:

```bash
curl -c cookies.txt -X POST http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@workspace.local","password":"ChangeMeAdmin1"}'
```

Access token is in the JSON body. Refresh token is the `refresh_token` HttpOnly cookie (path `/auth`).

## Tests

From this directory:

```bash
go test ./...
```

On Windows, if test binaries are blocked in `%TEMP%` (Application Control), run:

```bash
mkdir -p tmp
GOTMPDIR="$(pwd)/tmp" go test ./...
```

Unit tests cover login/refresh/logout, RBAC 401/403, last Super Admin guard, and department name uniqueness. They do not require Postgres.

## Layout

`cmd/server` → `internal/{auth,user,department,rbac,seed,httpserver}` → `pkg/*`

Super Admin is seeded with **all permission codes** (no middleware bypass). Custom roles are deferred (BE-1.10).

`Dockerfile` is kept for a later production image; it is not used in local compose.
