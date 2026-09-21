package rbac

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"workspace/pkg/apperr"
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
	Name            string   `json:"name" validate:"required,min=2,max=128"`
	Code            string   `json:"code" validate:"omitempty,max=64"`
	PermissionCodes []string `json:"permissionCodes" validate:"required,min=1,dive,required"`
}

type patchRequest struct {
	Name            *string   `json:"name" validate:"omitempty,min=2,max=128"`
	PermissionCodes *[]string `json:"permissionCodes" validate:"omitempty,min=1,dive,required"`
}

type Public struct {
	ID              uuid.UUID `json:"id"`
	Code            string    `json:"code"`
	Name            string    `json:"name"`
	IsSystem        bool      `json:"isSystem"`
	PermissionCodes []string  `json:"permissionCodes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type PermissionPublic struct {
	ID   uuid.UUID `json:"id"`
	Code string    `json:"code"`
	Name string    `json:"name"`
}

func toPublic(role *Role) Public {
	codes := PermissionCodes(role)
	if codes == nil {
		codes = []string{}
	}
	return Public{
		ID: role.ID, Code: role.Code, Name: role.Name, IsSystem: role.IsSystem,
		PermissionCodes: codes, CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt,
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.List()
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]Public, 0, len(rows))
	for i := range rows {
		items = append(items, toPublic(&rows[i]))
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	role, err := h.svc.Get(id)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toPublic(role))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req createRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	role, err := h.svc.Create(CreateInput{
		Name: req.Name, Code: req.Code, PermissionCodes: req.PermissionCodes,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, toPublic(role))
}

func (h *Handler) Patch(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
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
	role, err := h.svc.Patch(id, PatchInput{Name: req.Name, PermissionCodes: req.PermissionCodes})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, toPublic(role))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListPermissions()
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]PermissionPublic, 0, len(rows))
	for _, p := range rows {
		items = append(items, PermissionPublic{ID: p.ID, Code: p.Code, Name: p.Name})
	}
	response.JSON(w, http.StatusOK, items)
}

func parseID(r *http.Request) (uuid.UUID, error) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, apperr.ErrValidation.WithMessage("invalid id")
	}
	return id, nil
}
