package migrations

import (
	"gorm.io/gorm"
)

func CreateUsersTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			email_verified_at TIMESTAMP NULL,
			password VARCHAR(255) NOT NULL,
			remember_token VARCHAR(100) NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL,
			deleted_at TIMESTAMP NULL,
			INDEX idx_users_email (email),
			INDEX idx_users_deleted_at (deleted_at)
		)
	`).Error
}

func DropUsersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS users").Error
}
