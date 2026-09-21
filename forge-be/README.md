# WorkSpace backend

Go API (chi, GORM, Zap, JWT + RBAC) at `forge/forge-be`.

```bash
cp .env.example .env
docker compose up -d
go run ./cmd/server
```



- Health: `GET http://localhost:8080/health`
- Ready: `GET http://localhost:8080/ready`
- Swagger: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

```bash
go test ./...
```

