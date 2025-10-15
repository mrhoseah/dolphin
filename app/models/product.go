package models

import (
	"time"
	"gorm.io/gorm"
)

// Product represents a product model
type Product struct {
	ID        uint           `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	// Add your fields here
	// Name string `gorm:"not null"`
	// Email string `gorm:"uniqueIndex"`
}

// TableName returns the table name for the product model
func (Product) TableName() string {
	return "product"
}

// BeforeCreate is called before creating a new record
func (m *Product) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	return nil
}

// BeforeUpdate is called before updating a record
func (m *Product) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *Product) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}
