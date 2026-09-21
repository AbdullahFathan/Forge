package resource

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"workspace/internal/auditlog"
	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/internal/task"
	"workspace/internal/user"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type Store interface {
	Create(*Allocation) error
	Save(*Allocation) error
	GetByID(uuid.UUID) (*Allocation, error)
	SoftDelete(uuid.UUID) error
	List(ListFilter) ([]Allocation, int64, error)
	OverlappingSameProject(userID, projectID uuid.UUID, from, to time.Time, excludeID *uuid.UUID) ([]Allocation, error)
	ForUsersInRange([]uuid.UUID, time.Time, time.Time) ([]Allocation, error)
	AllInRange(time.Time, time.Time) ([]Allocation, error)
	ListHolidays() ([]Holiday, error)
	CreateHoliday(*Holiday) error
	GetHolidayByDate(time.Time) (*Holiday, error)
	DeleteHoliday(uuid.UUID) error
	ListActiveUsers(*uuid.UUID, string) ([]user.User, error)
	GetUser(uuid.UUID) (*user.User, error)
}

type Projects interface {
	MustSee(actor authctx.Principal, projectID uuid.UUID) (*project.Project, string, error)
	MustManage(actor authctx.Principal, projectID uuid.UUID) (*project.Project, error)
	UpsertMemberRole(projectID, userID uuid.UUID, role string) error
}

type AssignedTasks interface {
	ListByAssignee(userID uuid.UUID) ([]task.Task, error)
}

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	repo  Store
	proj  Projects
	tasks AssignedTasks
	audit auditlog.Auditor
	clock Clock
}

func NewService(repo Store, proj Projects, tasks AssignedTasks, audit auditlog.Auditor, clock Clock) *Service {
	if audit == nil {
		audit = auditlog.Noop{}
	}
	if clock == nil {
		clock = realClock{}
	}
	return &Service{repo: repo, proj: proj, tasks: tasks, audit: audit, clock: clock}
}

type CreateInput struct {
	UserID            uuid.UUID
	ProjectID         uuid.UUID
	AllocationPercent float64
	StartDate         time.Time
	EndDate           time.Time
	Role              string
}

type PatchInput struct {
	AllocationPercent *float64
	StartDate         *time.Time
	EndDate           *time.Time
	Role              *string
}

func (s *Service) scopeAll(actor authctx.Principal) bool {
	return authctx.HasPermission(actor, perm.ProjectReadAll)
}

func (s *Service) authorizeWrite(actor authctx.Principal, projectID uuid.UUID) error {
	if s.scopeAll(actor) {
		_, _, err := s.proj.MustSee(actor, projectID)
		return err
	}
	_, err := s.proj.MustManage(actor, projectID)
	return err
}

func (s *Service) List(actor authctx.Principal, f ListFilter) ([]Allocation, int64, error) {
	f.ActorID = actor.UserID
	f.ScopeAll = s.scopeAll(actor)
	return s.repo.List(f)
}

func (s *Service) Get(actor authctx.Principal, id uuid.UUID) (*Allocation, []string, bool, error) {
	a, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, false, err
	}
	if err := s.authorizeWrite(actor, a.ProjectID); err != nil {
		// list-style read: PM who cannot manage still forbidden; RM uses MustSee via scopeAll
		if !s.scopeAll(actor) {
			return nil, nil, false, err
		}
	}
	warn, over, err := s.userOverFlags(a.UserID, a.StartDate, a.EndDate)
	return a, warn, over, err
}

func (s *Service) Create(ctx context.Context, actor authctx.Principal, ip string, in CreateInput) (*Allocation, []string, bool, error) {
	if err := validateAlloc(in.AllocationPercent, in.StartDate, in.EndDate, in.Role); err != nil {
		return nil, nil, false, err
	}
	if in.Role == "" {
		in.Role = RoleMember
	}
	if err := s.authorizeWrite(actor, in.ProjectID); err != nil {
		return nil, nil, false, err
	}
	if _, err := s.repo.GetUser(in.UserID); err != nil {
		return nil, nil, false, apperr.ErrValidation.WithMessage("invalid userId")
	}
	overlap, err := s.repo.OverlappingSameProject(in.UserID, in.ProjectID, in.StartDate, in.EndDate, nil)
	if err != nil {
		return nil, nil, false, err
	}
	if len(overlap) > 0 {
		return nil, nil, false, apperr.ErrConflict.WithMessage("overlapping allocation for the same user and project")
	}
	a := &Allocation{
		UserID: in.UserID, ProjectID: in.ProjectID,
		AllocationPercent: in.AllocationPercent,
		StartDate:         DateUTC(in.StartDate), EndDate: DateUTC(in.EndDate),
		Role: in.Role,
	}
	if err := s.repo.Create(a); err != nil {
		return nil, nil, false, err
	}
	if err := s.proj.UpsertMemberRole(in.ProjectID, in.UserID, in.Role); err != nil {
		return nil, nil, false, err
	}
	got, err := s.repo.GetByID(a.ID)
	if err != nil {
		return nil, nil, false, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "ResourceAllocation", EntityID: got.ID,
		Action: "CREATED", After: got,
	})
	warn, over, err := s.userOverFlags(got.UserID, got.StartDate, got.EndDate)
	return got, warn, over, err
}

func (s *Service) Patch(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID, in PatchInput) (*Allocation, []string, bool, error) {
	a, err := s.repo.GetByID(id)
	if err != nil {
		return nil, nil, false, err
	}
	if err := s.authorizeWrite(actor, a.ProjectID); err != nil {
		return nil, nil, false, err
	}
	before := *a
	if in.AllocationPercent != nil {
		a.AllocationPercent = *in.AllocationPercent
	}
	if in.StartDate != nil {
		a.StartDate = DateUTC(*in.StartDate)
	}
	if in.EndDate != nil {
		a.EndDate = DateUTC(*in.EndDate)
	}
	if in.Role != nil {
		a.Role = *in.Role
	}
	if err := validateAlloc(a.AllocationPercent, a.StartDate, a.EndDate, a.Role); err != nil {
		return nil, nil, false, err
	}
	overlap, err := s.repo.OverlappingSameProject(a.UserID, a.ProjectID, a.StartDate, a.EndDate, &a.ID)
	if err != nil {
		return nil, nil, false, err
	}
	if len(overlap) > 0 {
		return nil, nil, false, apperr.ErrConflict.WithMessage("overlapping allocation for the same user and project")
	}
	if err := s.repo.Save(a); err != nil {
		return nil, nil, false, err
	}
	if in.Role != nil {
		_ = s.proj.UpsertMemberRole(a.ProjectID, a.UserID, a.Role)
	}
	got, err := s.repo.GetByID(a.ID)
	if err != nil {
		return nil, nil, false, err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "ResourceAllocation", EntityID: got.ID,
		Action: "UPDATED", Before: before, After: got,
	})
	warn, over, err := s.userOverFlags(got.UserID, got.StartDate, got.EndDate)
	return got, warn, over, err
}

func (s *Service) Delete(ctx context.Context, actor authctx.Principal, ip string, id uuid.UUID) error {
	a, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if err := s.authorizeWrite(actor, a.ProjectID); err != nil {
		return err
	}
	if err := s.repo.SoftDelete(id); err != nil {
		return err
	}
	_ = s.audit.Record(ctx, auditlog.Record{
		ActorID: actor.UserID, IP: ip, EntityType: "ResourceAllocation", EntityID: a.ID,
		Action: "DELETED", Before: a,
	})
	return nil
}

func (s *Service) userOverFlags(userID uuid.UUID, from, to time.Time) ([]string, bool, error) {
	rows, err := s.repo.ForUsersInRange([]uuid.UUID{userID}, from, to)
	if err != nil {
		return nil, false, err
	}
	daily := DailyPercent(toSlices(rows), from, to)
	if OverAllocated(daily) {
		return []string{WarningOverAllocated}, true, nil
	}
	return []string{}, false, nil
}

func toSlices(rows []Allocation) []Slice {
	out := make([]Slice, 0, len(rows))
	for _, a := range rows {
		out = append(out, Slice{
			UserID: a.UserID, ProjectID: a.ProjectID,
			Percent: a.AllocationPercent, Start: a.StartDate, End: a.EndDate,
		})
	}
	return out
}

func holidayMap(rows []Holiday) map[time.Time]struct{} {
	dates := make([]time.Time, 0, len(rows))
	for _, h := range rows {
		dates = append(dates, h.Date)
	}
	return HolidaySet(dates)
}

type RangeQuery struct {
	From         time.Time
	To           time.Time
	Granularity  string
	DepartmentID *uuid.UUID
	Skill        string
}

func (s *Service) parseRange(from, to time.Time, granularity string) (time.Time, time.Time, string, error) {
	from, to = DateUTC(from), DateUTC(to)
	if to.Before(from) {
		return time.Time{}, time.Time{}, "", apperr.ErrValidation.WithMessage("to must be on or after from")
	}
	days := InclusiveDays(from, to)
	if days < 28 || days > 84 {
		return time.Time{}, time.Time{}, "", apperr.ErrValidation.WithMessage("range must be between 4 and 12 weeks inclusive")
	}
	if granularity == "" {
		granularity = "week"
	}
	if granularity != "week" && granularity != "month" {
		return time.Time{}, time.Time{}, "", apperr.ErrValidation.WithMessage("granularity must be week or month")
	}
	return from, to, granularity, nil
}

func (s *Service) Forecast(q RangeQuery) (Forecast, error) {
	from, to, gran, err := s.parseRange(q.From, q.To, q.Granularity)
	if err != nil {
		return Forecast{}, err
	}
	users, err := s.repo.ListActiveUsers(q.DepartmentID, q.Skill)
	if err != nil {
		return Forecast{}, err
	}
	ids := make([]uuid.UUID, 0, len(users))
	for i := range users {
		ids = append(ids, users[i].ID)
	}
	allocs, err := s.repo.ForUsersInRange(ids, from, to)
	if err != nil {
		return Forecast{}, err
	}
	holidays, err := s.repo.ListHolidays()
	if err != nil {
		return Forecast{}, err
	}
	hset := holidayMap(holidays)
	periods := Periods(from, to, gran)
	byUser := map[uuid.UUID][]Allocation{}
	for _, a := range allocs {
		byUser[a.UserID] = append(byUser[a.UserID], a)
	}
	out := Forecast{From: dateStr(from), To: dateStr(to), Granularity: gran, Users: make([]UserForecast, 0, len(users))}
	for i := range users {
		u := users[i]
		out.Users = append(out.Users, UserForecast{
			UserID: u.ID, Name: u.Name, DepartmentID: u.DepartmentID,
			CapacityHoursPerDay: u.CapacityHoursPerDay,
			Skills:              user.SkillNames(&u),
			Buckets:             s.bucketsFor(u, byUser[u.ID], periods, hset),
		})
	}
	return out, nil
}

func (s *Service) bucketsFor(u user.User, allocs []Allocation, periods []Period, hset map[time.Time]struct{}) []Bucket {
	out := make([]Bucket, 0, len(periods))
	hours := u.CapacityHoursPerDay
	for _, p := range periods {
		daily := DailyPercent(toSlices(allocs), p.Start, p.End)
		allocH := AllocatedHours(daily, hours, hset)
		eff := EffectiveHours(p.Start, p.End, hours, hset)
		util := UtilizationPercent(allocH, eff)
		bd := map[uuid.UUID]Breakdown{}
		for _, a := range allocs {
			if !Overlaps(a.StartDate, a.EndDate, p.Start, p.End) {
				continue
			}
			item := bd[a.ProjectID]
			item.ProjectID = a.ProjectID
			item.AllocationPercent += a.AllocationPercent
			if a.Project != nil {
				item.ProjectName = a.Project.Name
			}
			bd[a.ProjectID] = item
		}
		keys := make([]uuid.UUID, 0, len(bd))
		for id := range bd {
			keys = append(keys, id)
		}
		sort.Slice(keys, func(i, j int) bool { return keys[i].String() < keys[j].String() })
		list := make([]Breakdown, 0, len(keys))
		for _, id := range keys {
			list = append(list, bd[id])
		}
		out = append(out, Bucket{
			PeriodKey: p.Key, PeriodStart: dateStr(p.Start), PeriodEnd: dateStr(p.End),
			UtilizationPercent: util, Band: Band(util),
			AllocatedHours: allocH, EffectiveHours: eff, AllocationBreakdown: list,
		})
	}
	return out
}

func (s *Service) Matrix(q RangeQuery) (Matrix, error) {
	fc, err := s.Forecast(q)
	if err != nil {
		return Matrix{}, err
	}
	if q.DepartmentID == nil && len(fc.Users) > matrixUserThreshold {
		return Matrix{}, apperr.ErrValidation.WithMessage("departmentId is required when more than 200 users match")
	}
	m := Matrix{
		Users:   make([]MatrixUser, 0, len(fc.Users)),
		Periods: nil,
		Cells:   map[string]map[string]MatrixCell{},
	}
	if len(fc.Users) > 0 {
		for _, b := range fc.Users[0].Buckets {
			m.Periods = append(m.Periods, MatrixPeriod{Key: b.PeriodKey, Start: b.PeriodStart, End: b.PeriodEnd})
		}
	} else {
		from, to, gran, _ := s.parseRange(q.From, q.To, q.Granularity)
		for _, p := range Periods(from, to, gran) {
			m.Periods = append(m.Periods, MatrixPeriod{Key: p.Key, Start: dateStr(p.Start), End: dateStr(p.End)})
		}
	}
	for _, u := range fc.Users {
		m.Users = append(m.Users, MatrixUser{ID: u.UserID, Name: u.Name})
		cells := map[string]MatrixCell{}
		for _, b := range u.Buckets {
			cells[b.PeriodKey] = MatrixCell{UtilizationPercent: b.UtilizationPercent, Band: b.Band}
		}
		m.Cells[u.UserID.String()] = cells
	}
	return m, nil
}

func (s *Service) Availability(q RangeQuery) ([]AvailabilityItem, error) {
	fc, err := s.Forecast(q)
	if err != nil {
		return nil, err
	}
	var out []AvailabilityItem
	for _, u := range fc.Users {
		var alloc, eff float64
		for _, b := range u.Buckets {
			alloc += b.AllocatedHours
			eff += b.EffectiveHours
		}
		util := UtilizationPercent(alloc, eff)
		if util >= 80 {
			continue
		}
		out = append(out, AvailabilityItem{
			UserID: u.UserID, Name: u.Name, DepartmentID: u.DepartmentID,
			UtilizationPercent: util, RemainingPercent: 100 - util, Band: Band(util),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].RemainingPercent > out[j].RemainingPercent })
	return out, nil
}

func (s *Service) OverloadAlerts() ([]OverloadAlert, error) {
	now := DateUTC(s.clock.Now())
	from, to := now, now.AddDate(0, 0, 13) // 14 days inclusive
	users, err := s.repo.ListActiveUsers(nil, "")
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(users))
	byID := map[uuid.UUID]user.User{}
	for i := range users {
		ids = append(ids, users[i].ID)
		byID[users[i].ID] = users[i]
	}
	allocs, err := s.repo.ForUsersInRange(ids, from, to)
	if err != nil {
		return nil, err
	}
	grouped := map[uuid.UUID][]Allocation{}
	for _, a := range allocs {
		grouped[a.UserID] = append(grouped[a.UserID], a)
	}
	var out []OverloadAlert
	for uid, rows := range grouped {
		daily := DailyPercent(toSlices(rows), from, to)
		if !OverAllocated(daily) {
			continue
		}
		u := byID[uid]
		out = append(out, OverloadAlert{
			UserID: uid, Name: u.Name, MaxPercent: MaxDailyPercent(daily),
			From: dateStr(from), To: dateStr(to),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MaxPercent > out[j].MaxPercent })
	return out, nil
}

func (s *Service) Workload(actor authctx.Principal, userID *uuid.UUID) (Workload, error) {
	target := actor.UserID
	if userID != nil && *userID != actor.UserID {
		if !authctx.HasPermission(actor, perm.UserManage) && !authctx.HasPermission(actor, perm.CapacityView) {
			return Workload{}, apperr.ErrForbidden
		}
		target = *userID
	}
	u, err := s.repo.GetUser(target)
	if err != nil {
		return Workload{}, err
	}
	now := DateUTC(s.clock.Now())
	to := now.AddDate(0, 0, 27) // 4 weeks inclusive
	rows, err := s.repo.ForUsersInRange([]uuid.UUID{target}, now, to)
	if err != nil {
		return Workload{}, err
	}
	holidays, err := s.repo.ListHolidays()
	if err != nil {
		return Workload{}, err
	}
	hset := holidayMap(holidays)
	periods := Periods(now, to, "week")
	wl := Workload{
		UserID:     target,
		Projects:   []WorkloadProject{},
		Tasks:      []WorkloadTask{},
		WeekSeries: s.bucketsFor(*u, rows, periods, hset),
	}
	for _, a := range rows {
		name := ""
		if a.Project != nil {
			name = a.Project.Name
		}
		wl.Projects = append(wl.Projects, WorkloadProject{
			ProjectID: a.ProjectID, ProjectName: name, AllocationPercent: a.AllocationPercent,
			StartDate: dateStr(a.StartDate), EndDate: dateStr(a.EndDate), Role: a.Role,
		})
	}
	if s.tasks != nil {
		tasks, err := s.tasks.ListByAssignee(target)
		if err != nil {
			return Workload{}, err
		}
		for i := range tasks {
			t := tasks[i]
			item := WorkloadTask{
				ID: t.ID, ProjectID: t.ProjectID, Name: t.Name, Status: t.Status, Priority: t.Priority,
			}
			if t.DueDate != nil {
				s := dateStr(*t.DueDate)
				item.DueDate = &s
			}
			wl.Tasks = append(wl.Tasks, item)
		}
	}
	return wl, nil
}

func (s *Service) ListHolidays() ([]Holiday, error) {
	return s.repo.ListHolidays()
}

func (s *Service) CreateHoliday(date time.Time, name string) (*Holiday, error) {
	date = DateUTC(date)
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperr.ErrValidation.WithMessage("name is required")
	}
	if _, err := s.repo.GetHolidayByDate(date); err == nil {
		return nil, apperr.ErrConflict.WithMessage("holiday already exists on this date")
	} else if ae, ok := apperr.As(err); !ok || ae.Code != apperr.ErrNotFound.Code {
		return nil, err
	}
	h := &Holiday{Date: date, Name: name}
	if err := s.repo.CreateHoliday(h); err != nil {
		return nil, err
	}
	return h, nil
}

func (s *Service) DeleteHoliday(id uuid.UUID) error {
	return s.repo.DeleteHoliday(id)
}

func validateAlloc(pct float64, start, end time.Time, role string) error {
	if pct < 0 || pct > 100 {
		return apperr.ErrValidation.WithMessage("allocationPercent must be between 0 and 100")
	}
	if DateUTC(end).Before(DateUTC(start)) {
		return apperr.ErrValidation.WithMessage("endDate must be on or after startDate")
	}
	if role != "" && role != RoleLead && role != RoleMember && role != RoleViewer {
		return apperr.ErrValidation.WithMessage("invalid project role")
	}
	return nil
}
