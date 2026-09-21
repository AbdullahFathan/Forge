package perm

const (
	UserManage       = "user.manage"
	RoleManage       = "role.manage"
	DepartmentManage = "department.manage"
	ProjectCreate    = "project.create"
	ProjectDelete    = "project.delete"
	ProjectReadAll   = "project.read_all"
	TaskManage       = "task.manage"
	ResourceAllocate = "resource.allocate"
	CapacityView     = "capacity.view"
	ReportExport     = "report.export"
	AuditRead        = "audit.read"
	SystemConfigure  = "system.configure"
)

func All() []string {
	return []string{
		UserManage,
		RoleManage,
		DepartmentManage,
		ProjectCreate,
		ProjectDelete,
		ProjectReadAll,
		TaskManage,
		ResourceAllocate,
		CapacityView,
		ReportExport,
		AuditRead,
		SystemConfigure,
	}
}

const (
	RoleSuperAdmin      = "SUPER_ADMIN"
	RoleAdmin           = "ADMIN"
	RoleResourceManager = "RESOURCE_MANAGER"
	RoleProjectManager  = "PROJECT_MANAGER"
	RoleMember          = "MEMBER"
)
