package server

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"workspace/config"
	"workspace/internal/auditlog"
	"workspace/internal/auth"
	"workspace/internal/dashboard"
	"workspace/internal/department"
	"workspace/internal/jobs"
	"workspace/internal/notification"
	"workspace/internal/project"
	"workspace/internal/rbac"
	"workspace/internal/rbac/perm"
	"workspace/internal/report"
	"workspace/internal/resource"
	"workspace/internal/task"
	"workspace/internal/user"
	"workspace/pkg/middleware"
	"workspace/pkg/response"
	"workspace/pkg/storage"
)

type App struct {
	Handler   http.Handler
	Scheduler *jobs.Scheduler
}

func NewRouter(cfg *config.Config, log *zap.Logger, db *gorm.DB, rdb *redis.Client, store storage.Client) *App {
	userRepo := user.NewRepository(db)
	deptRepo := department.NewRepository(db)
	roleRepo := rbac.NewRepository(db)
	auditSvc := auditlog.NewService(db)
	userSvc := user.NewService(userRepo, roleRepo, deptRepo, auditSvc)
	notifySvc := notification.NewService(db, mailSender(cfg, userRepo), userSvc)
	deptSvc := department.NewService(deptRepo)
	roleSvc := rbac.NewService(roleRepo)
	roleH := rbac.NewHandler(roleSvc)
	taskRepo := task.NewRepository(db)
	projRepo := project.NewRepository(db)
	projSvc := project.NewService(projRepo, userRepo, deptRepo, taskRepo, auditSvc)
	projSvc.SetActivity(auditSvc)
	projSvc.SetNotify(notifySvc)
	taskSvc := task.NewService(taskRepo, projSvc, auditSvc)
	taskSvc.SetNotify(notifySvc)
	tokenStore := auth.NewRedisStore(rdb)
	authSvc := auth.NewService(userRepo, tokenStore, cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)

	authH := auth.NewHandler(authSvc, cfg.CookieSecure, cfg.JWTRefreshTTL)
	userH := user.NewHandler(userSvc)
	deptH := department.NewHandler(deptSvc)
	projH := project.NewHandler(projSvc)
	taskH := task.NewHandler(taskSvc)
	resRepo := resource.NewRepository(db)
	resSvc := resource.NewService(resRepo, projSvc, taskRepo, auditSvc, nil)
	resH := resource.NewHandler(resSvc)
	auditH := auditlog.NewHandler(auditSvc)
	notifyH := notification.NewHandler(notifySvc)
	reportSvc := report.New(projSvc, projRepo, taskRepo, resSvc, userRepo, store)
	reportH := report.NewHandler(reportSvc)
	dashSvc := dashboard.New(projSvc, projRepo, taskRepo, resSvc, auditSvc, notifySvc, nil)
	dashH := dashboard.NewHandler(dashSvc)
	sched := jobs.New(nil, rdb, notifySvc, taskRepo, projRepo, resSvc, roleRepo)

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(corsHandler(cfg.CORSOrigins))
	r.Use(middleware.RequestLogger(log))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	r.Get("/ready", Ready(db, rdb))
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

		r.With(middleware.Require(perm.RoleManage)).Get("/permissions", roleH.ListPermissions)
		r.Route("/roles", func(r chi.Router) {
			r.Use(middleware.Require(perm.RoleManage))
			r.Get("/", roleH.List)
			r.Post("/", roleH.Create)
			r.Get("/{id}", roleH.Get)
			r.Patch("/{id}", roleH.Patch)
			r.Delete("/{id}", roleH.Delete)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Get("/", projH.List)
			r.With(middleware.Require(perm.ProjectCreate)).Post("/", projH.Create)
			r.Get("/{id}", projH.Get)
			r.Get("/{id}/activity", projH.Activity)
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

		r.Get("/me/workload", resH.Workload)

		r.Route("/resources", func(r chi.Router) {
			r.Group(func(r chi.Router) {
				r.Use(middleware.Require(perm.ResourceAllocate))
				r.Get("/allocations", resH.ListAllocations)
				r.Post("/allocations", resH.CreateAllocation)
				r.Patch("/allocations/{id}", resH.PatchAllocation)
				r.Delete("/allocations/{id}", resH.DeleteAllocation)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.Require(perm.CapacityView))
				r.Get("/capacity", resH.Capacity)
				r.Get("/matrix", resH.Matrix)
				r.Get("/availability", resH.Availability)
				r.Get("/overload-alerts", resH.OverloadAlerts)
			})
			r.Group(func(r chi.Router) {
				r.Use(middleware.Require(perm.DepartmentManage))
				r.Get("/holidays", resH.ListHolidays)
				r.Post("/holidays", resH.CreateHoliday)
				r.Patch("/holidays/{id}", resH.PatchHoliday)
				r.Delete("/holidays/{id}", resH.DeleteHoliday)
			})
		})

		r.With(middleware.Require(perm.AuditRead)).Get("/audit-logs", auditH.List)

		r.Route("/notifications", func(r chi.Router) {
			r.Get("/", notifyH.List)
			r.Get("/unread-count", notifyH.UnreadCount)
			r.Patch("/{id}/read", notifyH.MarkRead)
			r.Post("/read-all", notifyH.MarkAllRead)
		})

		r.Route("/reports", func(r chi.Router) {
			r.Use(middleware.Require(perm.ReportExport))
			r.Get("/project-status", reportH.ProjectStatus)
			r.Get("/resource-utilization", reportH.Utilization)
			r.Get("/task-completion", reportH.TaskCompletion)
		})

		r.Route("/dashboards", func(r chi.Router) {
			r.Get("/executive", dashH.Executive)
			r.Get("/project-manager", dashH.ProjectManager)
			r.Get("/member", dashH.Member)
		})
	})

	return &App{Handler: r, Scheduler: sched}
}

func mailSender(cfg *config.Config, users *user.Repository) notification.Sender {
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return notification.NopSender{}
	}
	return notification.NewSMTP(notification.SMTPConfig{
		Host:     cfg.SMTPHost,
		Port:     cfg.SMTPPort,
		Username: cfg.SMTPUser,
		Password: cfg.SMTPPassword,
		From:     cfg.SMTPFrom,
		StartTLS: cfg.SMTPStartTLS,
	}, users)
}
