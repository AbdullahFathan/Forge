# WorkSpace

WorkSpace is a project and resource management system. Teams plan work, assign people, track tasks, watch capacity, and export reports from one place.

The product is two apps:

- **`forge-be`** — Go HTTP API (JWT + RBAC). Users, departments, projects, tasks, allocations, capacity, notifications, reports, and audit logs. PostgreSQL for data, Redis for refresh tokens, object storage for report files.
- **`forge-fe`** — React UI. Sign-in, a role-aware dashboard, project and task boards, workload and capacity views, notifications, reports, and audit.

Access is permission-based (admin, project manager, resource manager, member, and similar roles). The UI only shows what the signed-in user’s claims allow.

## Local run

API first, then the UI.

```bash
cd forge-be
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

```bash
cd forge-fe
cp .env.example .env
bun install
bun run dev
```

- UI: http://localhost:3000
- API: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html

Point `VITE_API_URL` at the API. Include the Vite origin in the API `CORS_ORIGINS` so cookie-based refresh works.

Details for each app live in `forge-be/README.md` and `forge-fe/README.md`.
