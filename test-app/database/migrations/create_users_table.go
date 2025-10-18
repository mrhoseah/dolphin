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
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NULL,
			updated_at TIMESTAMP NULL
		)
	`).Error
}

func DropUsersTable(db *gorm.DB) error {
	return db.Exec("DROP TABLE IF EXISTS users").Error
}
