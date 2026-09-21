package project

import (
	"workspace/internal/rbac/perm"
	"workspace/pkg/authctx"
)

func CanSee(p authctx.Principal, pr *Project, memberRole string, isMember bool) bool {
	if authctx.HasPermission(p, perm.ProjectReadAll) {
		return true
	}
	if pr.OwnerID == p.UserID {
		return true
	}
	return isMember
}

func CanManage(p authctx.Principal, pr *Project, memberRole string) bool {
	if p.RoleCode == perm.RoleSuperAdmin || p.RoleCode == perm.RoleAdmin {
		return true
	}
	if pr.OwnerID == p.UserID {
		return true
	}
	return memberRole == RoleLead
}

func IsViewer(memberRole string) bool {
	return memberRole == RoleViewer
}
