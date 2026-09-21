package migrate

import (
	"fmt"

	"gorm.io/gorm"

	"workspace/internal/department"
	"workspace/internal/rbac"
	"workspace/internal/user"
)

func Auto(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&rbac.Permission{},
		&rbac.Role{},
		&department.Department{},
		&user.User{},
	); err != nil {
		return fmt.Errorf("automigrate: %w", err)
	}
	return nil
}
