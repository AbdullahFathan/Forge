package resource

import (
	"time"

	"github.com/google/uuid"
)

type UserBrief struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	DepartmentID   *uuid.UUID `json:"departmentId,omitempty"`
	DepartmentName *string    `json:"departmentName,omitempty"`
}

type Public struct {
	ID                uuid.UUID `json:"id"`
	UserID            uuid.UUID `json:"userId"`
	User              UserBrief `json:"user"`
	ProjectID         uuid.UUID `json:"projectId"`
	ProjectName       string    `json:"projectName,omitempty"`
	AllocationPercent float64   `json:"allocationPercent"`
	StartDate         string    `json:"startDate"`
	EndDate           string    `json:"endDate"`
	Role              string    `json:"role"`
	Warnings          []string  `json:"warnings"`
	OverAllocated     bool      `json:"overAllocated"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}

type HolidayPublic struct {
	ID        uuid.UUID `json:"id"`
	Date      string    `json:"date"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Breakdown struct {
	ProjectID         uuid.UUID `json:"projectId"`
	ProjectName       string    `json:"projectName,omitempty"`
	AllocationPercent float64   `json:"allocationPercent"`
}

type Bucket struct {
	PeriodKey           string      `json:"periodKey"`
	PeriodStart         string      `json:"periodStart"`
	PeriodEnd           string      `json:"periodEnd"`
	UtilizationPercent  float64     `json:"utilizationPercent"`
	Band                string      `json:"band"`
	AllocatedHours      float64     `json:"allocatedHours"`
	EffectiveHours      float64     `json:"effectiveHours"`
	AllocationBreakdown []Breakdown `json:"allocationBreakdown"`
}

type UserForecast struct {
	UserID              uuid.UUID  `json:"userId"`
	Name                string     `json:"name"`
	DepartmentID        *uuid.UUID `json:"departmentId,omitempty"`
	CapacityHoursPerDay int        `json:"capacityHoursPerDay"`
	Skills              []string   `json:"skills"`
	Buckets             []Bucket   `json:"buckets"`
}

type Forecast struct {
	From        string         `json:"from"`
	To          string         `json:"to"`
	Granularity string         `json:"granularity"`
	Users       []UserForecast `json:"users"`
}

type MatrixPeriod struct {
	Key   string `json:"key"`
	Start string `json:"start"`
	End   string `json:"end"`
}

type MatrixUser struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

type MatrixCell struct {
	UtilizationPercent float64 `json:"utilizationPercent"`
	Band               string  `json:"band"`
}

type Matrix struct {
	Users   []MatrixUser                     `json:"users"`
	Periods []MatrixPeriod                   `json:"periods"`
	Cells   map[string]map[string]MatrixCell `json:"cells"`
}

type AvailabilityItem struct {
	UserID             uuid.UUID  `json:"userId"`
	Name               string     `json:"name"`
	DepartmentID       *uuid.UUID `json:"departmentId,omitempty"`
	UtilizationPercent float64    `json:"utilizationPercent"`
	RemainingPercent   float64    `json:"remainingPercent"`
	Band               string     `json:"band"`
}

type OverloadAlert struct {
	UserID     uuid.UUID `json:"userId"`
	Name       string    `json:"name"`
	MaxPercent float64   `json:"maxPercent"`
	From       string    `json:"from"`
	To         string    `json:"to"`
}

type WorkloadProject struct {
	ProjectID         uuid.UUID `json:"projectId"`
	ProjectName       string    `json:"projectName,omitempty"`
	AllocationPercent float64   `json:"allocationPercent"`
	StartDate         string    `json:"startDate"`
	EndDate           string    `json:"endDate"`
	Role              string    `json:"role"`
}

type WorkloadTask struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"projectId"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Priority  string    `json:"priority"`
	DueDate   *string   `json:"dueDate"`
}

type Workload struct {
	UserID     uuid.UUID         `json:"userId"`
	Projects   []WorkloadProject `json:"projects"`
	Tasks      []WorkloadTask    `json:"tasks"`
	WeekSeries []Bucket          `json:"weekSeries"`
}

func dateStr(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return DateUTC(t).Format("2006-01-02")
}

func ToPublic(a *Allocation, warnings []string, over bool) Public {
	if warnings == nil {
		warnings = []string{}
	}
	out := Public{
		ID: a.ID, UserID: a.UserID, ProjectID: a.ProjectID,
		AllocationPercent: a.AllocationPercent,
		StartDate:         dateStr(a.StartDate), EndDate: dateStr(a.EndDate),
		Role: a.Role, Warnings: warnings, OverAllocated: over,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
		User: UserBrief{ID: a.UserID},
	}
	if a.User != nil {
		out.User.Name = a.User.Name
		out.User.DepartmentID = a.User.DepartmentID
		if a.User.Department != nil {
			out.User.DepartmentName = &a.User.Department.Name
		}
	}
	if a.Project != nil {
		out.ProjectName = a.Project.Name
	}
	return out
}

func ToHolidayPublic(h *Holiday) HolidayPublic {
	return HolidayPublic{ID: h.ID, Date: dateStr(h.Date), Name: h.Name, CreatedAt: h.CreatedAt}
}
