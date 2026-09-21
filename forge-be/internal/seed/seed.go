package seed

import (
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"workspace/internal/rbac"
	"workspace/internal/rbac/perm"
	"workspace/internal/resource"
	"workspace/internal/user"
)

type seedPerm struct {
	Code string
	Name string
}

func Run(db *gorm.DB, adminEmail, adminPassword string) error {
	perms := []seedPerm{
		{perm.UserManage, "Manage users and roles"},
		{perm.RoleManage, "Manage custom roles"},
		{perm.DepartmentManage, "Manage departments"},
		{perm.ProjectCreate, "Create projects"},
		{perm.ProjectDelete, "Delete or archive projects"},
		{perm.ProjectReadAll, "Read all projects"},
		{perm.TaskManage, "Manage tasks"},
		{perm.ResourceAllocate, "Allocate resources"},
		{perm.CapacityView, "View capacity planning"},
		{perm.ReportExport, "Export reports"},
		{perm.AuditRead, "Read audit logs"},
		{perm.SystemConfigure, "Configure system"},
	}

	permByCode := map[string]rbac.Permission{}
	for _, p := range perms {
		var row rbac.Permission
		if err := db.Where("code = ?", p.Code).FirstOrCreate(&row, rbac.Permission{Code: p.Code, Name: p.Name}).Error; err != nil {
			return fmt.Errorf("seed permission %s: %w", p.Code, err)
		}
		permByCode[p.Code] = row
	}

	pick := func(codes ...string) []rbac.Permission {
		out := make([]rbac.Permission, 0, len(codes))
		for _, c := range codes {
			out = append(out, permByCode[c])
		}
		return out
	}

	roles := []struct {
		Code        string
		Name        string
		Permissions []rbac.Permission
	}{
		{perm.RoleSuperAdmin, "Super Admin", pick(perm.All()...)},
		{perm.RoleAdmin, "Admin", pick(
			perm.UserManage, perm.RoleManage, perm.DepartmentManage,
			perm.ProjectCreate, perm.ProjectDelete, perm.ProjectReadAll, perm.TaskManage,
			perm.ResourceAllocate, perm.CapacityView, perm.ReportExport, perm.AuditRead,
		)},
		{perm.RoleResourceManager, "Resource Manager", pick(
			perm.ProjectReadAll, perm.ResourceAllocate, perm.CapacityView, perm.ReportExport,
		)},
		{perm.RoleProjectManager, "Project Manager", pick(
			perm.ProjectCreate, perm.ProjectDelete, perm.TaskManage, perm.ResourceAllocate, perm.ReportExport,
		)},
		{perm.RoleMember, "Member", pick(perm.TaskManage)},
	}

	var superAdmin rbac.Role
	for _, r := range roles {
		var role rbac.Role
		if err := db.Where("code = ?", r.Code).FirstOrCreate(&role, rbac.Role{Code: r.Code, Name: r.Name, IsSystem: true}).Error; err != nil {
			return fmt.Errorf("seed role %s: %w", r.Code, err)
		}
		if err := db.Model(&role).Association("Permissions").Replace(r.Permissions); err != nil {
			return fmt.Errorf("seed role permissions %s: %w", r.Code, err)
		}
		if r.Code == perm.RoleSuperAdmin {
			if err := db.Preload("Permissions").Where("code = ?", r.Code).First(&superAdmin).Error; err != nil {
				return err
			}
		}
	}

	if err := seedHolidays(db); err != nil {
		return err
	}

	var count int64
	if err := db.Model(&user.User{}).Where("email = ?", adminEmail).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), 12)
	if err != nil {
		return err
	}
	admin := user.User{
		Name:         "Super Admin",
		Email:        adminEmail,
		PasswordHash: string(hash),
		RoleID:       superAdmin.ID,
		IsActive:     true,
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("seed bootstrap admin: %w", err)
	}
	return nil
}

func seedHolidays(db *gorm.DB) error {
	items := []struct {
		date string
		name string
	}{
		{"2026-01-01", "Tahun Baru Masehi"},
		{"2026-03-21", "Hari Suci Nyepi"},
		{"2026-05-01", "Hari Buruh Internasional"},
		{"2026-08-17", "Hari Kemerdekaan RI"},
		{"2026-12-25", "Hari Natal"},
	}
	for _, it := range items {
		d, err := time.Parse("2006-01-02", it.date)
		if err != nil {
			return err
		}
		var n int64
		if err := db.Model(&resource.Holiday{}).Where("date = ?", d).Count(&n).Error; err != nil {
			return fmt.Errorf("seed holiday %s: %w", it.date, err)
		}
		if n > 0 {
			continue
		}
		h := resource.Holiday{Date: d, Name: it.name}
		if err := db.Create(&h).Error; err != nil {
			return fmt.Errorf("seed holiday %s: %w", it.date, err)
		}
	}
	return nil
}
