package auditlog

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"workspace/pkg/apperr"
	"workspace/pkg/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f, err := parseFilter(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	rows, total, err := h.svc.List(f)
	if err != nil {
		response.Error(w, err)
		return
	}
	if r.URL.Query().Get("format") == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", `attachment; filename="audit-logs.csv"`)
		_ = WriteCSV(w, rows)
		return
	}
	items := make([]Public, 0, len(rows))
	for _, row := range rows {
		items = append(items, ToPublic(row))
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total})
}

func parseFilter(r *http.Request) (ListFilter, error) {
	f := ListFilter{
		Page:       queryInt(r, "page", 1),
		PageSize:   queryInt(r, "pageSize", 50),
		EntityType: r.URL.Query().Get("entityType"),
		Action:     r.URL.Query().Get("action"),
	}
	if v := r.URL.Query().Get("userId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, apperr.ErrValidation.WithMessage("invalid userId")
		}
		f.UserID = &id
	}
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse("2006-01-02", v)
			if err != nil {
				return f, apperr.ErrValidation.WithMessage("from must be RFC3339 or YYYY-MM-DD")
			}
		}
		f.From = &t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			t, err = time.Parse("2006-01-02", v)
			if err != nil {
				return f, apperr.ErrValidation.WithMessage("to must be RFC3339 or YYYY-MM-DD")
			}
		}
		f.To = &t
	}
	return f, nil
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
