package resource

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
	UserID            uuid.UUID `json:"userId" validate:"required"`
	ProjectID         uuid.UUID `json:"projectId" validate:"required"`
	AllocationPercent float64   `json:"allocationPercent" validate:"gte=0,lte=100"`
	StartDate         string    `json:"startDate" validate:"required"`
	EndDate           string    `json:"endDate" validate:"required"`
	Role              string    `json:"role" validate:"omitempty,oneof=LEAD MEMBER VIEWER"`
}

type patchRequest struct {
	AllocationPercent *float64 `json:"allocationPercent" validate:"omitempty,gte=0,lte=100"`
	StartDate         *string  `json:"startDate"`
	EndDate           *string  `json:"endDate"`
	Role              *string  `json:"role" validate:"omitempty,oneof=LEAD MEMBER VIEWER"`
}

type holidayRequest struct {
	Date string `json:"date" validate:"required"`
	Name string `json:"name" validate:"required,min=1,max=128"`
}

func (h *Handler) ListAllocations(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	f := ListFilter{
		Page:     queryInt(r, "page", 1),
		PageSize: queryInt(r, "pageSize", 20),
	}
	if v := r.URL.Query().Get("userId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid userId"))
			return
		}
		f.UserID = &id
	}
	if v := r.URL.Query().Get("projectId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid projectId"))
			return
		}
		f.ProjectID = &id
	}
	from, err := queryDate(r, "from")
	if err != nil {
		response.Error(w, err)
		return
	}
	to, err := queryDate(r, "to")
	if err != nil {
		response.Error(w, err)
		return
	}
	f.From, f.To = from, to
	rows, total, err := h.svc.List(actor, f)
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]Public, 0, len(rows))
	for i := range rows {
		warn, over, _ := h.svc.userOverFlags(rows[i].UserID, rows[i].StartDate, rows[i].EndDate)
		items = append(items, ToPublic(&rows[i], warn, over))
	}
	response.JSON(w, http.StatusOK, Page{Items: items, Page: f.Page, PageSize: f.PageSize, TotalItems: total})
}

func (h *Handler) CreateAllocation(w http.ResponseWriter, r *http.Request) {
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
	end, err := parseDate(req.EndDate)
	if err != nil {
		response.Error(w, err)
		return
	}
	a, warn, over, err := h.svc.Create(r.Context(), actor, clientIP(r), CreateInput{
		UserID: req.UserID, ProjectID: req.ProjectID, AllocationPercent: req.AllocationPercent,
		StartDate: start, EndDate: end, Role: req.Role,
	})
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToPublic(a, warn, over))
}

func (h *Handler) PatchAllocation(w http.ResponseWriter, r *http.Request) {
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
	in := PatchInput{AllocationPercent: req.AllocationPercent, Role: req.Role}
	if req.StartDate != nil {
		d, err := parseDate(*req.StartDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		in.StartDate = &d
	}
	if req.EndDate != nil {
		d, err := parseDate(*req.EndDate)
		if err != nil {
			response.Error(w, err)
			return
		}
		in.EndDate = &d
	}
	a, warn, over, err := h.svc.Patch(r.Context(), actor, clientIP(r), id, in)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, ToPublic(a, warn, over))
}

func (h *Handler) DeleteAllocation(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Capacity(w http.ResponseWriter, r *http.Request) {
	q, err := parseRangeQuery(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	fc, err := h.svc.Forecast(q)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, fc)
}

func (h *Handler) Matrix(w http.ResponseWriter, r *http.Request) {
	q, err := parseRangeQuery(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	m, err := h.svc.Matrix(q)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, m)
}

func (h *Handler) Availability(w http.ResponseWriter, r *http.Request) {
	q, err := parseRangeQuery(r)
	if err != nil {
		response.Error(w, err)
		return
	}
	items, err := h.svc.Availability(q)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) OverloadAlerts(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.OverloadAlerts()
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) Workload(w http.ResponseWriter, r *http.Request) {
	actor, ok := authctx.User(r.Context())
	if !ok {
		response.Error(w, apperr.ErrUnauthorized)
		return
	}
	var uid *uuid.UUID
	if v := r.URL.Query().Get("userId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			response.Error(w, apperr.ErrValidation.WithMessage("invalid userId"))
			return
		}
		uid = &id
	}
	wl, err := h.svc.Workload(actor, uid)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, wl)
}

func (h *Handler) ListHolidays(w http.ResponseWriter, r *http.Request) {
	rows, err := h.svc.ListHolidays()
	if err != nil {
		response.Error(w, err)
		return
	}
	items := make([]HolidayPublic, 0, len(rows))
	for i := range rows {
		items = append(items, ToHolidayPublic(&rows[i]))
	}
	response.JSON(w, http.StatusOK, items)
}

func (h *Handler) CreateHoliday(w http.ResponseWriter, r *http.Request) {
	var req holidayRequest
	if err := response.Decode(r, &req); err != nil {
		response.Error(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		response.Error(w, err)
		return
	}
	d, err := parseDate(req.Date)
	if err != nil {
		response.Error(w, err)
		return
	}
	hday, err := h.svc.CreateHoliday(d, req.Name)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusCreated, ToHolidayPublic(hday))
}

func (h *Handler) DeleteHoliday(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r, "id")
	if err != nil {
		response.Error(w, err)
		return
	}
	if err := h.svc.DeleteHoliday(id); err != nil {
		response.Error(w, err)
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func parseRangeQuery(r *http.Request) (RangeQuery, error) {
	fromS := strings.TrimSpace(r.URL.Query().Get("from"))
	toS := strings.TrimSpace(r.URL.Query().Get("to"))
	if fromS == "" || toS == "" {
		return RangeQuery{}, apperr.ErrValidation.WithMessage("from and to are required (YYYY-MM-DD)")
	}
	from, err := parseDate(fromS)
	if err != nil {
		return RangeQuery{}, err
	}
	to, err := parseDate(toS)
	if err != nil {
		return RangeQuery{}, err
	}
	q := RangeQuery{From: from, To: to, Granularity: strings.TrimSpace(r.URL.Query().Get("granularity")), Skill: strings.TrimSpace(r.URL.Query().Get("skill"))}
	if v := r.URL.Query().Get("departmentId"); v != "" {
		id, err := uuid.Parse(v)
		if err != nil {
			return RangeQuery{}, apperr.ErrValidation.WithMessage("invalid departmentId")
		}
		q.DepartmentID = &id
	}
	return q, nil
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

func queryDate(r *http.Request, key string) (*time.Time, error) {
	v := strings.TrimSpace(r.URL.Query().Get(key))
	if v == "" {
		return nil, nil
	}
	t, err := parseDate(v)
	if err != nil {
		return nil, err
	}
	return &t, nil
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
