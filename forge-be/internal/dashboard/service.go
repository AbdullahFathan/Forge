package dashboard

import (
	"sort"
	"time"

	"github.com/google/uuid"

	"workspace/internal/auditlog"
	"workspace/internal/notification"
	"workspace/internal/project"
	"workspace/internal/rbac/perm"
	"workspace/internal/resource"
	"workspace/internal/task"
	"workspace/pkg/apperr"
	"workspace/pkg/authctx"
)

type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now().UTC() }

type Service struct {
	clock    Clock
	projects *project.Service
	projRepo *project.Repository
	tasks    *task.Repository
	res      *resource.Service
	audit    *auditlog.Service
	notify   *notification.Service
}

func New(projects *project.Service, projRepo *project.Repository, tasks *task.Repository, res *resource.Service, audit *auditlog.Service, notify *notification.Service, clock Clock) *Service {
	if clock == nil {
		clock = realClock{}
	}
	return &Service{clock: clock, projects: projects, projRepo: projRepo, tasks: tasks, res: res, audit: audit, notify: notify}
}

type MonthPoint struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type Executive struct {
	ActiveProjects    int64                       `json:"activeProjects"`
	CompletedProjects int64                       `json:"completedProjects"`
	LateProjects      int64                       `json:"lateProjects"`
	OrgUtilization    float64                     `json:"orgUtilization"`
	DueSoon           []project.Public            `json:"dueSoon"`
	TopUtilization    []resource.AvailabilityItem `json:"topUtilization"`
	MonthlyCompleted  []MonthPoint                `json:"monthlyCompleted"`
}

type PMDash struct {
	Projects     []project.Public            `json:"projects"`
	Blocked      []task.Public               `json:"blocked"`
	Overdue      []task.Public               `json:"overdue"`
	Unassigned   []task.Public               `json:"unassigned"`
	TeamCapacity []resource.AvailabilityItem `json:"teamCapacity"`
	Activity     []auditlog.Public           `json:"recentActivity"`
}

type MemberDash struct {
	Tasks         []task.Public         `json:"tasks"`
	WeekLoad      resource.Workload     `json:"weekLoad"`
	Notifications []notification.Public `json:"notifications"`
}

func (s *Service) Executive(actor authctx.Principal) (Executive, error) {
	if actor.RoleCode != perm.RoleAdmin && actor.RoleCode != perm.RoleSuperAdmin {
		return Executive{}, apperr.ErrForbidden
	}
	today := dateUTC(s.clock.Now())
	counts, err := s.projRepo.StatusCounts(today)
	if err != nil {
		return Executive{}, err
	}
	due, err := s.projRepo.DueBetween(today.AddDate(0, 0, 7), today.AddDate(0, 0, 14))
	if err != nil {
		return Executive{}, err
	}
	duePub := make([]project.Public, 0, len(due))
	for i := range due {
		pct, _, _ := s.tasks.Completion(due[i].ID)
		duePub = append(duePub, project.ToPublic(&due[i], pct))
	}
	weekFrom, weekTo := today, today.AddDate(0, 0, 6)
	utils, err := s.res.PeriodUtilization(weekFrom, weekTo, nil)
	if err != nil {
		return Executive{}, err
	}
	var sum float64
	for _, u := range utils {
		sum += u.UtilizationPercent
	}
	org := 0.0
	if len(utils) > 0 {
		org = sum / float64(len(utils))
	}
	top := utils
	if len(top) > 5 {
		top = top[:5]
	}
	months, err := s.projRepo.CompletedByMonth(today.AddDate(0, -5, 0))
	if err != nil {
		return Executive{}, err
	}
	trend := make([]MonthPoint, 0, len(months))
	for _, m := range months {
		trend = append(trend, MonthPoint{Month: m.Month, Count: m.Count})
	}
	return Executive{
		ActiveProjects: counts.Active, CompletedProjects: counts.Done, LateProjects: counts.Late,
		OrgUtilization: org, DueSoon: duePub, TopUtilization: top, MonthlyCompleted: trend,
	}, nil
}

func (s *Service) ProjectManager(actor authctx.Principal) (PMDash, error) {
	if actor.RoleCode != perm.RoleProjectManager && actor.RoleCode != perm.RoleAdmin && actor.RoleCode != perm.RoleSuperAdmin {
		return PMDash{}, apperr.ErrForbidden
	}
	today := dateUTC(s.clock.Now())
	owned, err := s.projRepo.ListOwned(actor.UserID)
	if err != nil {
		return PMDash{}, err
	}
	if actor.RoleCode == perm.RoleAdmin || actor.RoleCode == perm.RoleSuperAdmin {
		owned, err = s.projRepo.ListAllActive()
		if err != nil {
			return PMDash{}, err
		}
	}
	pubs := make([]project.Public, 0, len(owned))
	ids := make([]uuid.UUID, 0, len(owned))
	memberSet := map[uuid.UUID]struct{}{}
	for i := range owned {
		pct, _, _ := s.tasks.Completion(owned[i].ID)
		pubs = append(pubs, project.ToPublic(&owned[i], pct))
		ids = append(ids, owned[i].ID)
		mids, _ := s.projRepo.MemberUserIDs(owned[i].ID)
		for _, id := range mids {
			memberSet[id] = struct{}{}
		}
	}
	att, err := s.tasks.Attention(ids, today)
	if err != nil {
		return PMDash{}, err
	}
	weekFrom, weekTo := today, today.AddDate(0, 0, 6)
	utils, err := s.res.PeriodUtilization(weekFrom, weekTo, nil)
	if err != nil {
		return PMDash{}, err
	}
	var team []resource.AvailabilityItem
	for _, u := range utils {
		if _, ok := memberSet[u.UserID]; ok {
			team = append(team, u)
		}
	}
	sort.Slice(team, func(i, j int) bool { return team[i].UtilizationPercent > team[j].UtilizationPercent })
	var act []auditlog.Public
	if s.audit != nil && len(ids) > 0 {
		rows, _, _ := s.audit.List(auditlog.ListFilter{ProjectID: &ids[0], Page: 1, PageSize: 15})
		for _, row := range rows {
			act = append(act, auditlog.ToPublic(row))
		}
	}
	return PMDash{
		Projects: pubs,
		Blocked:  mapTasks(att.Blocked), Overdue: mapTasks(att.Overdue), Unassigned: mapTasks(att.Unassigned),
		TeamCapacity: team, Activity: act,
	}, nil
}

func (s *Service) Member(actor authctx.Principal) (MemberDash, error) {
	rows, err := s.tasks.ListByAssignee(actor.UserID)
	if err != nil {
		return MemberDash{}, err
	}
	wl, err := s.res.Workload(actor, nil)
	if err != nil {
		return MemberDash{}, err
	}
	var notes []notification.Public
	if s.notify != nil {
		ns, _ := s.notify.Latest(actor.UserID, 10)
		for _, n := range ns {
			notes = append(notes, notification.ToPublic(n))
		}
	}
	nearest := rows
	if len(nearest) > 20 {
		nearest = nearest[:20]
	}
	return MemberDash{Tasks: mapTasks(nearest), WeekLoad: wl, Notifications: notes}, nil
}

func mapTasks(rows []task.Task) []task.Public {
	out := make([]task.Public, 0, len(rows))
	for i := range rows {
		out = append(out, task.ToPublic(&rows[i], false))
	}
	return out
}

func dateUTC(t time.Time) time.Time {
	y, m, d := t.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
