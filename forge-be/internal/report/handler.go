package report

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) ProjectStatus(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f, err := parseFilter(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	rows, err := h.svc.ProjectStatus(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	headers := []string{"id", "name", "status", "completionPercent", "targetEndDate", "ownerId", "ownerName"}
	var csvRows [][]string
	for _, x := range rows {
		csvRows = append(csvRows, []string{x.ID.String(), x.Name, x.Status, fmtPct(x.CompletionPercent), x.TargetEndDate, x.OwnerID.String(), x.OwnerName})
	}
	h.respond(w, r, "project-status", headers, csvRows, rows)
}

func (h *Handler) Utilization(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f, err := parseFilter(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	rows, err := h.svc.Utilization(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	headers := []string{"userId", "name", "utilizationPercent", "band"}
	var csvRows [][]string
	for _, x := range rows {
		csvRows = append(csvRows, []string{x.UserID.String(), x.Name, fmtPct(x.UtilizationPercent), x.Band})
	}
	h.respond(w, r, "resource-utilization", headers, csvRows, rows)
}

func (h *Handler) TaskCompletion(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f, err := parseFilter(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	rows, err := h.svc.TaskCompletion(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	headers := []string{"id", "name", "done", "total", "percent"}
	var csvRows [][]string
	for _, x := range rows {
		csvRows = append(csvRows, []string{x.ID.String(), x.Name, strconv.FormatInt(x.Done, 10), strconv.FormatInt(x.Total, 10), fmtPct(x.Percent)})
	}
	h.respond(w, r, "task-completion", headers, csvRows, rows)
}

func (h *Handler) respond(w http.ResponseWriter, r *http.Request, name string, headers []string, csvRows [][]string, jsonBody any) {
	format := r.URL.Query().Get("format")
	if format == "" || format == "json" {
		response.JSON(w, http.StatusOK, jsonBody)
		return
	}
	if format == "csv" {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.csv"`, name))
		_ = WriteCSV(w, headers, csvRows)
		return
	}
	if format == "pdf" {
		pdfHeaders, pdfRows := headers, csvRows
		if name == "project-status" {
			pdfHeaders, pdfRows = omitColumns(headers, csvRows, "id", "ownerId")
		}
		body, err := WritePDF(name, pdfHeaders, pdfRows)
		if err != nil {
			response.Error(w, err)
			return
		}
		key := fmt.Sprintf("reports/%s-%d.pdf", name, time.Now().Unix())
		url, _ := h.svc.MaybeStore(key, body, "application/pdf")
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, name))
		if url != "" {
			w.Header().Set("X-Object-URL", url)
		}
		_, _ = w.Write(body)
		return
	}
	response.Error(w, apperr.ErrValidation.WithMessage("format must be json, csv, or pdf"))
}

func parseFilter(r *http.Request) (Filter, error) {
	f := Filter{GroupBy: r.URL.Query().Get("groupBy")}
	if v := r.URL.Query().Get("departmentId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, apperr.ErrValidation.WithMessage("invalid departmentId")
		}
		f.DepartmentID = &id
	}
	if v := r.URL.Query().Get("projectId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return f, apperr.ErrValidation.WithMessage("invalid projectId")
		}
		f.ProjectID = &id
	}
	if v := r.URL.Query().Get("from"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, apperr.ErrValidation.WithMessage("from must be YYYY-MM-DD")
		}
		f.From = &t
	}
	if v := r.URL.Query().Get("to"); v != "" {
		t, err := time.Parse("2006-01-02", v)
		if err != nil {
			return f, apperr.ErrValidation.WithMessage("to must be YYYY-MM-DD")
		}
		f.To = &t
	}
	return f, nil
}
