package project

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

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
	Name          string     `json:"name" validate:"required,min=3,max=200"`
	Description   string     `json:"description" validate:"omitempty,max=4000"`
	Status        string     `json:"status" validate:"omitempty,oneof=DRAFT ACTIVE ON_HOLD COMPLETED ARCHIVED"`
	Priority      string     `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	StartDate     string     `json:"startDate" validate:"required"`
	TargetEndDate string     `json:"targetEndDate" validate:"required"`
	OwnerID       uuid.UUID  `json:"ownerId" validate:"required"`
	DepartmentID  *uuid.UUID `json:"departmentId"`
	Tags          []string   `json:"tags"`
}

type patchRequest struct {
	Name            *string    `json:"name" validate:"omitempty,min=3,max=200"`
	Description     *string    `json:"description" validate:"omitempty,max=4000"`
	Status          *string    `json:"status" validate:"omitempty,oneof=DRAFT ACTIVE ON_HOLD COMPLETED ARCHIVED"`
	Priority        *string    `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	StartDate       *string    `json:"startDate"`
	TargetEndDate   *string    `json:"targetEndDate"`
	OwnerID         *uuid.UUID `json:"ownerId"`
	DepartmentID    *uuid.UUID `json:"departmentId"`
	ClearDepartment *bool      `json:"clearDepartment"`
	Tags            *[]string  `json:"tags"`
}

type memberRequest struct {
	UserID uuid.UUID `json:"userId" validate:"required"`
	Role   string    `json:"role" validate:"omitempty,oneof=LEAD MEMBER VIEWER"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f := ListFilter{
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "pageSize", 20),
		Status:   strings.TrimSpace(r.URL.Query().Get("status")),
	}
	if v := r.URL.Query().Get("departmentId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid departmentId"))
			return
		}
		f.DepartmentID = &id
	}
	rows, percents, total, err := h.svc.List(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]Public, 0, len(rows))
	for i := range rows {
		items = append(items, ToPublic(&rows[i], percents[i]))
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
	start, err := parseDate(req.StartDate)
	if err != nil {
		response.Error(w, err)
		return
	}
	end, err := parseDate(req.TargetEndDate)
	if err != nil {
		response.Error(w, err)
		return
	}
	p, err := h.svc.Create(r.Context(), actor, clientIP(r), CreateInput{
		Name: req.Name, Description: req.Description, Status: req.Status, Priority: req.Priority,
		StartDate: start, TargetEndDate: end, OwnerID: req.OwnerID, DepartmentID: req.DepartmentID, Tags: req.Tags,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	pct, _, _ := h.svc.completion(p.ID)
	response.JSON(w, http.StatusCreated, ToPublic(p, pct))
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	_, sum, err := h.svc.Get(actor, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, sum)
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	var req patchRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	in := PatchInput{
		Name: req.Name, Description: req.Description, Status: req.Status, Priority: req.Priority,
		OwnerID: req.OwnerID, Tags: req.Tags,
	}
	if req.StartDate != nil {
		d, err := parseDate(*req.StartDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		in.StartDate = &d
	}
	if req.TargetEndDate != nil {
		d, err := parseDate(*req.TargetEndDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		in.TargetEndDate = &d
	}
	if req.ClearDepartment != nil && *req.ClearDepartment {
		var nilID *uuid.UUID
		in.DepartmentID = &nilID
	} else if req.DepartmentID != nil {
		idCopy := req.DepartmentID
		in.DepartmentID = &idCopy
	}
	p, err := h.svc.Patch(r.Context(), actor, clientIP(r), id, in)
	if err != nil {
		response.Error(w, err)
		return
	}
	pct, _, _ := h.svc.completion(p.ID)
	response.JSON(w, http.StatusOK, ToPublic(p, pct))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.svc.Delete(r.Context(), actor, clientIP(r), id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListMembers(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	page := queryInt(r, "page", 1)
	pageSize := queryInt(r, "pageSize", 20)
	rows, total, err := h.svc.ListMembers(actor, id, page, pageSize)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]MemberPublic, 0, len(rows))
	for i := range rows {
		items = append(items, ToMemberPublic(&rows[i]))
	}
	response.JSON(w, http.StatusOK, MemberPage{Items: items, Page: page, PageSize: pageSize, TotalItems: total})
}

func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	var req memberRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	m, err := h.svc.AddMember(r.Context(), actor, clientIP(r), id, req.UserID, req.Role)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToMemberPublic(m))
}

func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	uid, err := parseID(r, "userId")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.svc.RemoveMember(r.Context(), actor, clientIP(r), id, uid); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, name))
	if err != nil {
		return uuid.Nil, apperr.ErrValidation.WithMessage("invalid " + name)
	}
	return id, nil
}

func parseDate(s string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Time{}, apperr.ErrValidation.WithMessage("dates must be YYYY-MM-DD")
	}
	return t, nil
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
