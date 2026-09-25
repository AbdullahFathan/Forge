package report

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/phpdave11/gofpdf"

	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/internal/resource"
	"workspace/internal/task"
	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
	"workspace/pkg/storage"
)

type Service struct {
	projects *project.Service
	projRepo *project.Repository
	tasks    *task.Repository
	res      *resource.Service
	users    *user.Repository
	store    storage.Client
}

func New(projects *project.Service, projRepo *project.Repository, tasks *task.Repository, res *resource.Service, users *user.Repository, store storage.Client) *Service {
	if store == nil {
		store = storage.Nop{}
	}
	return &Service{projects: projects, projRepo: projRepo, tasks: tasks, res: res, users: users, store: store}
}

type Filter struct {
	From         *time.Time
	To           *time.Time
	DepartmentID *uuid.UUID
	ProjectID    *uuid.UUID
	GroupBy      string
}

type ProjectStatusRow struct {
	ID                uuid.UUID `json:"id"`
	Name              string    `json:"name"`
	Status            string    `json:"status"`
	CompletionPercent float64   `json:"completionPercent"`
	TargetEndDate     string    `json:"targetEndDate"`
	OwnerID           uuid.UUID `json:"ownerId"`
	OwnerName         string    `json:"ownerName,omitempty"`
}

type UtilizationRow struct {
	UserID             uuid.UUID `json:"userId"`
	Name               string    `json:"name"`
	UtilizationPercent float64   `json:"utilizationPercent"`
	Band               string    `json:"band"`
}

type CompletionRow struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	Done    int64     `json:"done"`
	Total   int64     `json:"total"`
	Percent float64   `json:"percent"`
}

func (s *Service) scopedProjectIDs(actor authctx.Principal, f Filter) ([]uuid.UUID, []project.Project, error) {
	if !authctx.HasPermission(actor, perm.ReportExport) {
		return nil, nil, apperr.ErrForbidden
	}
	lf := project.ListFilter{Page: 1, PageSize: 100}
	if f.DepartmentID != nil {
		lf.DepartmentID = f.DepartmentID
	}
	rows, _, total, err := s.projects.List(actor, lf)
	if err != nil {
		return nil, nil, err
	}
	if total > 100 {
		lf.PageSize = int(total)
		if lf.PageSize > 2000 {
			lf.PageSize = 2000
		}
		rows, _, _, err = s.projects.List(actor, lf)
		if err != nil {
			return nil, nil, err
		}
	}
	var ids []uuid.UUID
	var out []project.Project
	for _, p := range rows {
		if f.ProjectID != nil && p.ID != *f.ProjectID {
			continue
		}
		if f.From != nil && p.TargetEndDate.Before(*f.From) {
			continue
		}
		if f.To != nil && p.TargetEndDate.After(*f.To) {
			continue
		}
		ids = append(ids, p.ID)
		out = append(out, p)
	}
	return ids, out, nil
}

func (s *Service) ProjectStatus(actor authctx.Principal, f Filter) ([]ProjectStatusRow, error) {
	_, rows, err := s.scopedProjectIDs(actor, f)
	if err != nil {
		return nil, err
	}
	out := make([]ProjectStatusRow, 0, len(rows))
	for i := range rows {
		p := rows[i]
		pct, _, _ := s.tasks.Completion(p.ID)
		name := ""
		if p.Owner != nil {
			name = p.Owner.Name
		}
		out = append(out, ProjectStatusRow{
			ID: p.ID, Name: p.Name, Status: p.Status, CompletionPercent: pct,
			TargetEndDate: p.TargetEndDate.UTC().Format("2006-01-02"),
			OwnerID:       p.OwnerID, OwnerName: name,
		})
	}
	return out, nil
}

func (s *Service) Utilization(actor authctx.Principal, f Filter) ([]UtilizationRow, error) {
	if !authctx.HasPermission(actor, perm.ReportExport) {
		return nil, apperr.ErrForbidden
	}
	from, to := time.Now().UTC().AddDate(0, 0, -28), time.Now().UTC()
	if f.From != nil {
		from = *f.From
	}
	if f.To != nil {
		to = *f.To
	}
	if !authctx.HasPermission(actor, perm.CapacityView) && !authctx.HasPermission(actor, perm.ProjectReadAll) {
		// PM: still allowed via ReportExport; utilization limited by department of owned work — use org util if RM/Admin else same engine scoped by department
		return s.utilFromCapacity(from, to, f.DepartmentID)
	}
	return s.utilFromCapacity(from, to, f.DepartmentID)
}

func (s *Service) utilFromCapacity(from, to time.Time, dept *uuid.UUID) ([]UtilizationRow, error) {
	items, err := s.res.PeriodUtilization(from, to, dept)
	if err != nil {
		return nil, err
	}
	out := make([]UtilizationRow, 0, len(items))
	for _, it := range items {
		out = append(out, UtilizationRow{
			UserID: it.UserID, Name: it.Name, UtilizationPercent: it.UtilizationPercent, Band: it.Band,
		})
	}
	return out, nil
}

func (s *Service) TaskCompletion(actor authctx.Principal, f Filter) ([]CompletionRow, error) {
	ids, projs, err := s.scopedProjectIDs(actor, f)
	if err != nil {
		return nil, err
	}
	group := f.GroupBy
	if group == "" {
		group = "project"
	}
	if group != "project" && group != "user" {
		return nil, apperr.ErrValidation.WithMessage("groupBy must be project or user")
	}
	if group == "project" {
		stats, err := s.tasks.CompletionByProject(ids)
		if err != nil {
			return nil, err
		}
		out := make([]CompletionRow, 0, len(projs))
		for _, p := range projs {
			st := stats[p.ID]
			pct := 0.0
			if st.Total > 0 {
				pct = float64(st.Done) / float64(st.Total) * 100
			}
			out = append(out, CompletionRow{ID: p.ID, Name: p.Name, Done: st.Done, Total: st.Total, Percent: pct})
		}
		return out, nil
	}
	stats, err := s.tasks.CompletionByUser(ids)
	if err != nil {
		return nil, err
	}
	out := make([]CompletionRow, 0, len(stats))
	for uid, st := range stats {
		name := uid.String()
		if u, err := s.users.GetByID(uid); err == nil {
			name = u.Name
		}
		pct := 0.0
		if st.Total > 0 {
			pct = float64(st.Done) / float64(st.Total) * 100
		}
		out = append(out, CompletionRow{ID: uid, Name: name, Done: st.Done, Total: st.Total, Percent: pct})
	}
	return out, nil
}

func WriteCSV(w io.Writer, headers []string, rows [][]string) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(headers); err != nil {
		return err
	}
	if err := cw.WriteAll(rows); err != nil {
		return err
	}
	cw.Flush()
	return cw.Error()
}

func omitColumns(headers []string, rows [][]string, names ...string) ([]string, [][]string) {
	skip := make(map[string]struct{}, len(names))
	for _, n := range names {
		skip[n] = struct{}{}
	}
	keep := make([]int, 0, len(headers))
	outH := make([]string, 0, len(headers))
	for i, h := range headers {
		if _, drop := skip[h]; drop {
			continue
		}
		keep = append(keep, i)
		outH = append(outH, h)
	}
	outR := make([][]string, len(rows))
	for r, row := range rows {
		nr := make([]string, 0, len(keep))
		for _, i := range keep {
			if i < len(row) {
				nr = append(nr, row[i])
			}
		}
		outR[r] = nr
	}
	return outH, outR
}

func WritePDF(title string, headers []string, rows [][]string) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetTitle(title, false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 10, title)
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 9)
	col := 270.0 / float64(len(headers))
	for _, h := range headers {
		pdf.CellFormat(col, 8, h, "1", 0, "", false, 0, "")
	}
	pdf.Ln(-1)
	for _, row := range rows {
		for _, c := range row {
			pdf.CellFormat(col, 7, c, "1", 0, "", false, 0, "")
		}
		pdf.Ln(-1)
	}
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *Service) MaybeStore(key string, body []byte, ct string) (string, error) {
	if err := s.store.Put(context.Background(), key, body, ct); err != nil {
		return "", err
	}
	return s.store.PresignGet(context.Background(), key, time.Hour)
}

func fmtPct(v float64) string { return fmt.Sprintf("%.1f", v) }
