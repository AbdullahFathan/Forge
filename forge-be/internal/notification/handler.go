package notification

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
	"workspace/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f := ListFilter{
		UserID:   actor.UserID,
		Type:     r.URL.Query().Get("type"),
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "pageSize", 20),
	}
	if v := r.URL.Query().Get("unread"); v == "true" || v == "1" {
		t := true
		f.Unread = &t
	} else if v == "false" || v == "0" {
		t := false
		f.Unread = &t
	}
	rows, total, err := h.svc.List(f)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]Public, 0, len(rows))
	for _, n := range rows {
		items = append(items, ToPublic(n))
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total})
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	n, err := h.svc.UnreadCount(actor.UserID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]int64{"unreadCount": n})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
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
	if err := h.svc.MarkRead(actor.UserID, id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	if err := h.svc.MarkAllRead(actor.UserID); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
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
