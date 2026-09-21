package dashboard

import (
	"net/http"

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

func (h *Handler) Executive(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	data, err := h.svc.Executive(actor)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (h *Handler) ProjectManager(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	data, err := h.svc.ProjectManager(actor)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

func (h *Handler) Member(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	data, err := h.svc.Member(actor)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, data)
}
