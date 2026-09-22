# WorkSpace API

Go HTTP API for **WorkSpace**, a project and resource management system. It handles authentication, users, departments, projects, tasks, capacity planning, notifications, reports, and audit logs.

The server is JWT + RBAC. Access tokens carry permission claims; refresh tokens are stored in Redis and issued as HttpOnly cookies. PostgreSQL holds application data. Object storage is used for report exports. Optional SMTP sends notification email.

## Stack

- Go, chi, GORM (PostgreSQL), Redis, Zap
- JWT auth with role-based permissions
- Docker Compose for Postgres, Redis, and object storage

## Run

```bash
cp .env.example .env
docker compose up -d
go run ./cmd/server
```

- Health: `GET http://localhost:8080/health`
- Ready: `GET http://localhost:8080/ready`
- Swagger: http://localhost:8080/swagger/index.html

The first start migrates the schema and seeds roles, permissions, and a bootstrap admin from env (`BOOTSTRAP_ADMIN_EMAIL`, `BOOTSTRAP_ADMIN_PASSWORD`).

```bash
go test ./...
```
