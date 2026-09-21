package user

import (
	"encoding/json"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
	"workspace/pkg/response"
	"workspace/pkg/validator"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Name                string     `json:"name" validate:"required,min=2,max=128"`
	Email               string     `json:"email" validate:"required,email"`
	Password            string     `json:"password" validate:"required,min=8"`
	RoleID              uuid.UUID  `json:"roleId" validate:"required"`
	DepartmentID        *uuid.UUID `json:"departmentId"`
	CapacityHoursPerDay int        `json:"capacityHoursPerDay" validate:"omitempty,min=1,max=24"`
	IsActive            *bool      `json:"isActive"`
	Skills              []string   `json:"skills"`
}

type patchRequest struct {
	Name                *string    `json:"name" validate:"omitempty,min=2,max=128"`
	Password            *string    `json:"password" validate:"omitempty,min=8"`
	RoleID              *uuid.UUID `json:"roleId"`
	DepartmentID        *uuid.UUID `json:"departmentId"`
	ClearDepartment     *bool      `json:"clearDepartment"`
	CapacityHoursPerDay *int       `json:"capacityHoursPerDay" validate:"omitempty,min=1,max=24"`
	IsActive            *bool      `json:"isActive"`
	Skills              *[]string  `json:"skills"`
}

type mePatchRequest struct {
	Name                      *string `json:"name" validate:"omitempty,min=2,max=128"`
	Password                  *string `json:"password" validate:"omitempty,min=8"`
	CurrentPassword           *string `json:"currentPassword"`
	EmailNotificationsEnabled *bool   `json:"emailNotificationsEnabled"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "pageSize", 20),
	}
	if v := r.URL.Query().Get("departmentId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid departmentId"))
			return
		}
		f.DepartmentID = &id
	}
	if v := r.URL.Query().Get("roleId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid roleId"))
			return
		}
		f.RoleID = &id
	}
	rows, total, err := h.svc.List(f)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]Public, 0, len(rows))
	for i := range rows {
		items = append(items, ToPublic(&rows[i]))
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	var req createRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	active := true
	if req.IsActive != nil {
		active = *req.IsActive
	}
	u, err := h.svc.Create(r.Context(), actor, clientIP(r), CreateInput{
		Name:                req.Name,
		Email:               req.Email,
		Password:            req.Password,
		RoleID:              req.RoleID,
		DepartmentID:        req.DepartmentID,
		CapacityHoursPerDay: req.CapacityHoursPerDay,
		IsActive:            active,
		Skills:              req.Skills,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToPublic(u))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperr.ErrValidation.WithMessage("invalid id"))
		return
	}
	u, err := h.svc.Get(id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(u))
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperr.ErrValidation.WithMessage("invalid id"))
		return
	}
	var req patchRequest
	if err := decodeAllowUnknown(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	in := PatchInput{
		Name:                req.Name,
		Password:            req.Password,
		RoleID:              req.RoleID,
		CapacityHoursPerDay: req.CapacityHoursPerDay,
		IsActive:            req.IsActive,
		Skills:              req.Skills,
	}
	if req.ClearDepartment != nil && *req.ClearDepartment {
		var nilID *uuid.UUID
		in.DepartmentID = &nilID
	} else if req.DepartmentID != nil {
		idCopy := req.DepartmentID
		in.DepartmentID = &idCopy
	}
	u, err := h.svc.Patch(r.Context(), actor, clientIP(r), id, in)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(u))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, apperr.ErrValidation.WithMessage("invalid id"))
		return
	}
	if err := h.svc.Delete(r.Context(), actor, clientIP(r), id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	p, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	u, err := h.svc.Get(p.UserID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(u))
}

func (h *Handler) PatchMe(w http.ResponseWriter, r *http.Request) {
	p, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	var req mePatchRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	u, err := h.svc.PatchMe(r.Context(), p, clientIP(r), p.UserID, MePatchInput{
		Name:                      req.Name,
		Password:                  req.Password,
		CurrentPassword:           req.CurrentPassword,
		EmailNotificationsEnabled: req.EmailNotificationsEnabled,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(u))
}

func queryInt(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func decodeAllowUnknown(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(dst); err != nil {
		return apperr.ErrValidation.WithMessage("invalid JSON body")
	}
	return nil
}
