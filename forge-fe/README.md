# WorkSpace frontend

Vite + React + TypeScript UI for **WorkSpace**, a project and resource management system. People sign in, then work from a dashboard through users, departments, projects, tasks, workload, capacity, notifications, reports, and audit.

The app talks to the WorkSpace HTTP API. Session access tokens stay in memory. Refresh uses an HttpOnly cookie (`withCredentials`), so the API CORS origins must include this app’s origin.

## Stack

- React 19, React Router, Vite, TypeScript
- TanStack Query, Axios, Zustand
- Tailwind CSS and shadcn/ui
- Vitest for unit tests

## Screens

- Login and profile
- Role-aware dashboard
- Users and departments (permission-gated)
- Projects, members, activity, and a task board
- My workload, resource allocations, and capacity views
- Notifications, report exports, and audit logs

## Run

```bash
cp .env.example .env
bun install
bun run dev
```

App: http://localhost:3000

Set `VITE_API_URL` to the API base URL (default `http://localhost:8080`).

```bash
bun run lint
bun run build
bun test
```
