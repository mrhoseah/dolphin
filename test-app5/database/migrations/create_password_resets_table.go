package migrations

import (
	"gorm.io/gorm"
)

func CreatePasswordResetsTable(db *gorm.DB) error {
	return db.Exec(`
		CREATE TABLE IF NOT EXISTS password_resets (
			id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
			email VARCHAR(255) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NULL,
			expires_at TIMESTAMP NULL,
			INDEX idx_password_resets_email (email),
			INDEX idx_password_resets_token (token)
		)
	`).Error
}

func DropPasswordResetsTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS password_resets").Error
}
