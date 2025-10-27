package models

import (
	"time"
	"gorm.io/gorm"
)

// Category represents a product category in the POS system
type Category struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	ParentID    *uint          `json:"parent_id"`
	Parent      *Category      `json:"parent" gorm:"foreignKey:ParentID"`
	Children    []Category     `json:"children" gorm:"foreignKey:ParentID"`
	Products    []Product      `json:"products" gorm:"foreignKey:CategoryID"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	SortOrder   int            `json:"sort_order" gorm:"default:0"`
	ImageURL    string         `json:"image_url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the category model
func (Category) TableName() string {
	return "categories"
}

// BeforeCreate is called before creating a new record
func (m *Category) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	return nil
}

// BeforeUpdate is called before updating a record
func (m *Category) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *Category) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// IsRoot checks if this is a root category
func (c *Category) IsRoot() bool {
	return c.ParentID == nil
}

// HasChildren checks if this category has subcategories
func (c *Category) HasChildren() bool {
	return len(c.Children) > 0
}

// GetProductCount returns the number of products in this category
func (c *Category) GetProductCount() int {
	return len(c.Products)
}
