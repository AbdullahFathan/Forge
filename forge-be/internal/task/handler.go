package task

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
	ParentTaskID   *uuid.UUID  `json:"parentTaskId"`
	Name           string      `json:"name" validate:"required,min=2,max=200"`
	Description    string      `json:"description" validate:"omitempty,max=8000"`
	Status         string      `json:"status" validate:"omitempty,oneof=BACKLOG TODO IN_PROGRESS IN_REVIEW DONE BLOCKED"`
	Priority       string      `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	EstimatedHours *float64    `json:"estimatedHours" validate:"omitempty,gte=0"`
	StartDate      *string     `json:"startDate"`
	DueDate        *string     `json:"dueDate"`
	Labels         []string    `json:"labels"`
	Position       *int        `json:"position"`
	AssigneeIDs    []uuid.UUID `json:"assigneeIds"`
}

type patchRequest struct {
	ParentTaskID   *uuid.UUID   `json:"parentTaskId"`
	ClearParent    *bool        `json:"clearParent"`
	Name           *string      `json:"name" validate:"omitempty,min=2,max=200"`
	Description    *string      `json:"description"`
	Status         *string      `json:"status" validate:"omitempty,oneof=BACKLOG TODO IN_PROGRESS IN_REVIEW DONE BLOCKED"`
	Priority       *string      `json:"priority" validate:"omitempty,oneof=LOW MEDIUM HIGH CRITICAL"`
	EstimatedHours *float64     `json:"estimatedHours"`
	ClearEstimate  *bool        `json:"clearEstimatedHours"`
	StartDate      *string      `json:"startDate"`
	ClearStartDate *bool        `json:"clearStartDate"`
	DueDate        *string      `json:"dueDate"`
	ClearDueDate   *bool        `json:"clearDueDate"`
	Labels         *[]string    `json:"labels"`
	Position       *int         `json:"position"`
	AssigneeIDs    *[]uuid.UUID `json:"assigneeIds"`
}

type depRequest struct {
	DependsOnTaskID uuid.UUID `json:"dependsOnTaskId" validate:"required"`
}

type commentRequest struct {
	Body string `json:"body" validate:"required,min=1,max=4000"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	pid, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	f := ListFilter{
		ProjectID: pid,
		Page:      queryInt(r, "page", 1),
		PageSize:  queryInt(r, "pageSize", 20),
		Status:    strings.TrimSpace(r.URL.Query().Get("status")),
		Priority:  strings.TrimSpace(r.URL.Query().Get("priority")),
		Label:     strings.TrimSpace(r.URL.Query().Get("label")),
		Include:   r.URL.Query().Get("include"),
	}
	if v := r.URL.Query().Get("assignee"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid assignee"))
			return
		}
		f.Assignee = &id
	}
	if v := r.URL.Query().Get("dueFrom"); v != "" {
		d, err := parseDate(v)
		if err != nil {
			response.Error(w, err)
			return
		}
		f.DueFrom = &d
	}
	if v := r.URL.Query().Get("dueTo"); v != "" {
		d, err := parseDate(v)
		if err != nil {
			response.Error(w, err)
			return
		}
		f.DueTo = &d
	}
	rows, total, err := h.svc.List(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	include := f.Include == "subtasks"
	items := make([]Public, 0, len(rows))
	for i := range rows {
		items = append(items, ToPublic(&rows[i], include))
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	pid, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
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
	start, err := optionalDate(req.StartDate)
	if err != nil {
		response.Error(w, err)
		return
	}
	due, err := optionalDate(req.DueDate)
	if err != nil {
		response.Error(w, err)
		return
	}
	t, err := h.svc.Create(r.Context(), actor, clientIP(r), CreateInput{
		ProjectID: pid, ParentTaskID: req.ParentTaskID, Name: req.Name, Description: req.Description,
		Status: req.Status, Priority: req.Priority, EstimatedHours: req.EstimatedHours,
		StartDate: start, DueDate: due, Labels: req.Labels, Position: req.Position, AssigneeIDs: req.AssigneeIDs,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToPublic(t, false))
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
	include := r.URL.Query().Get("include") == "subtasks"
	t, err := h.svc.Get(actor, id, include)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(t, include))
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
		Labels: req.Labels, Position: req.Position, AssigneeIDs: req.AssigneeIDs,
	}
	if req.ClearParent != nil && *req.ClearParent {
		var nilID *uuid.UUID
		in.ParentTaskID = &nilID
	} else if req.ParentTaskID != nil {
		idCopy := req.ParentTaskID
		in.ParentTaskID = &idCopy
	}
	if req.ClearEstimate != nil && *req.ClearEstimate {
		var n *float64
		in.EstimatedHours = &n
	} else if req.EstimatedHours != nil {
		v := req.EstimatedHours
		in.EstimatedHours = &v
	}
	if req.ClearStartDate != nil && *req.ClearStartDate {
		var n *time.Time
		in.StartDate = &n
	} else if req.StartDate != nil {
		d, err := parseDate(*req.StartDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		dp := &d
		in.StartDate = &dp
	}
	if req.ClearDueDate != nil && *req.ClearDueDate {
		var n *time.Time
		in.DueDate = &n
	} else if req.DueDate != nil {
		d, err := parseDate(*req.DueDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		dp := &d
		in.DueDate = &dp
	}
	t, err := h.svc.Patch(r.Context(), actor, clientIP(r), id, in)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(t, false))
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

func (h *Handler) AddDependency(w http.ResponseWriter, r *http.Request) {
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
	var req depRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	t, err := h.svc.AddDependency(r.Context(), actor, clientIP(r), id, req.DependsOnTaskID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToPublic(t, false))
}

func (h *Handler) RemoveDependency(w http.ResponseWriter, r *http.Request) {
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
	depID, err := parseID(r, "depId")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.svc.RemoveDependency(r.Context(), actor, clientIP(r), id, depID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListComments(w http.ResponseWriter, r *http.Request) {
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
	rows, err := h.svc.ListComments(actor, id)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]CommentPublic, 0, len(rows))
	for i := range rows {
		items = append(items, ToCommentPublic(&rows[i]))
	}
	response.JSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) AddComment(w http.ResponseWriter, r *http.Request) {
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
	var req commentRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	c, err := h.svc.AddComment(r.Context(), actor, clientIP(r), id, req.Body)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToCommentPublic(c))
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

func optionalDate(s *string) (*time.Time, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	d, err := parseDate(*s)
	if err != nil {
		return nil, err
	}
	return &d, nil
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
