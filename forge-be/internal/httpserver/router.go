package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"workspace/config"
	"workspace/internal/auditlog"
	"workspace/internal/auth"
	"workspace/internal/department"
	"workspace/internal/project"
	"workspace/internal/rbac"
	"workspace/internal/rbac/perm"
	"workspace/internal/task"
	"workspace/internal/user"
	"workspace/pkg/middleware"
	"workspace/pkg/response"
)

func NewRouter(cfg *config.Config, log *zap.Logger, db *gorm.DB, rdb *redis.Client) http.Handler {
	userRepo := user.NewRepository(db)
	deptRepo := department.NewRepository(db)
	roleRepo := rbac.NewRepository(db)

	userSvc := user.NewService(userRepo, roleRepo, deptRepo)
	deptSvc := department.NewService(deptRepo)
	taskRepo := task.NewRepository(db)
	projRepo := project.NewRepository(db)
	auditor := auditlog.Noop{}
	projSvc := project.NewService(projRepo, userRepo, deptRepo, taskRepo, auditor)
	taskSvc := task.NewService(taskRepo, projSvc, auditor)
	tokenStore := auth.NewRedisStore(rdb)
	authSvc := auth.NewService(userRepo, tokenStore, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	authH := auth.NewHandler(authSvc, cfg.CookieSecure, cfg.JWTRefreshTTL)
	userH := user.NewHandler(userSvc)
	deptH := department.NewHandler(deptSvc)
	projH := project.NewHandler(projSvc)
	taskH := task.NewHandler(taskSvc)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.CORSOrigins,
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(middleware.RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Handle("/swagger/openapi.yaml", SwaggerSpec())
	r.Get("/swagger", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/swagger/index.html", http.StatusFound)
	})
	r.Get("/swagger/index.html", SwaggerUI())

	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", authH.Login)
		r.Post("/refresh", authH.Refresh)
		r.Post("/logout", authH.Logout)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(cfg.JWTSecret))

		r.Route("/users", func(r chi.Router) {
			r.Get("/me", userH.GetMe)
			r.Patch("/me", userH.PatchMe)
			r.Group(func(r chi.Router) {
				r.Use(middleware.Require(perm.UserManage))
				r.Get("/", userH.List)
				r.Post("/", userH.Create)
				r.Get("/{id}", userH.Get)
				r.Patch("/{id}", userH.Patch)
				r.Delete("/{id}", userH.Delete)
			})
		})

		r.Route("/departments", func(r chi.Router) {
			r.Use(middleware.Require(perm.DepartmentManage))
			r.Get("/", deptH.List)
			r.Post("/", deptH.Create)
			r.Get("/{id}", deptH.Get)
			r.Patch("/{id}", deptH.Patch)
			r.Delete("/{id}", deptH.Delete)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projH.List)
			r.With(middleware.Require(perm.ProjectCreate)).Post("/", projH.Create)
			r.Get("/{id}", projH.Get)
			r.Patch("/{id}", projH.Patch)
			r.With(middleware.Require(perm.ProjectDelete)).Delete("/{id}", projH.Delete)
			r.Get("/{id}/members", projH.ListMembers)
			r.Post("/{id}/members", projH.AddMember)
			r.Delete("/{id}/members/{userId}", projH.RemoveMember)
			r.Get("/{id}/tasks", taskH.List)
			r.With(middleware.Require(perm.TaskManage)).Post("/{id}/tasks", taskH.Create)
		})

		r.Route("/tasks", func(r chi.Router) {
			r.Get("/{id}", taskH.Get)
			r.With(middleware.Require(perm.TaskManage)).Patch("/{id}", taskH.Patch)
			r.With(middleware.Require(perm.TaskManage)).Delete("/{id}", taskH.Delete)
			r.With(middleware.Require(perm.TaskManage)).Post("/{id}/dependencies", taskH.AddDependency)
			r.With(middleware.Require(perm.TaskManage)).Delete("/{id}/dependencies/{depId}", taskH.RemoveDependency)
			r.Get("/{id}/comments", taskH.ListComments)
			r.With(middleware.Require(perm.TaskManage)).Post("/{id}/comments", taskH.AddComment)
		})
	})

	return r
}
