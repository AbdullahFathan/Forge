package user

import (
	"time"

	"github.com/google/uuid"
)

type Public struct {
	ID                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Email               string     `json:"email"`
	RoleID              uuid.UUID  `json:"roleId"`
	RoleCode            string     `json:"roleCode"`
	DepartmentID        *uuid.UUID `json:"departmentId"`
	DepartmentName      *string    `json:"departmentName,omitempty"`
	CapacityHoursPerDay         int      `json:"capacityHoursPerDay"`
	IsActive                    bool     `json:"isActive"`
	EmailNotificationsEnabled   bool     `json:"emailNotificationsEnabled"`
	Skills                      []string `json:"skills"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

func ToPublic(u *User) Public {
	p := Public{
		ID:                  u.ID,
		Name:                u.Name,
		Email:               u.Email,
		RoleID:              u.RoleID,
		RoleCode:            u.Role.Code,
		DepartmentID:        u.DepartmentID,
		CapacityHoursPerDay:       u.CapacityHoursPerDay,
		IsActive:                  u.IsActive,
		EmailNotificationsEnabled: u.EmailNotificationsEnabled,
		Skills:                    SkillNames(u),
		CreatedAt:           u.CreatedAt,
		UpdatedAt:           u.UpdatedAt,
	}
	if u.Department != nil {
		p.DepartmentName = &u.Department.Name
	}
	return p
}

type Page struct {
	Items      []Public `json:"items"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalItems int64    `json:"totalItems"`
}
