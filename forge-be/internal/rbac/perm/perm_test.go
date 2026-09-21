package perm

import "testing"

func TestAllContainsRoleManage(t *testing.T) {
	found := false
	for _, c := range All() {
		if c == RoleManage {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected role.manage in All()")
	}
	if len(All()) < 10 {
		t.Fatalf("unexpected permission count %d", len(All()))
	}
}
