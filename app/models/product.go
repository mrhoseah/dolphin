package models

import (
	"time"
	"gorm.io/gorm"
)

// Product represents a product in the POS system
type Product struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	SKU         string         `json:"sku" gorm:"uniqueIndex"`
	Barcode     string         `json:"barcode" gorm:"uniqueIndex"`
	Price       float64        `json:"price" gorm:"not null;type:decimal(10,2)"`
	Cost        float64        `json:"cost" gorm:"type:decimal(10,2)"`
	CategoryID  uint           `json:"category_id"`
	Category    Category       `json:"category" gorm:"foreignKey:CategoryID"`
	Stock       int            `json:"stock" gorm:"default:0"`
	MinStock    int            `json:"min_stock" gorm:"default:0"`
	MaxStock    int            `json:"max_stock" gorm:"default:0"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	ImageURL    string         `json:"image_url"`
	TaxRate     float64        `json:"tax_rate" gorm:"default:0;type:decimal(5,2)"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the product model
func (Product) TableName() string {
	return "products"
}

// BeforeCreate is called before creating a new record
func (m *Product) BeforeCreate(tx *gorm.DB) error {
	// Generate SKU if not provided
	if m.SKU == "" {
		m.SKU = generateSKU(m.Name)
	}
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

// GetTotalPrice calculates the total price including tax
func (p *Product) GetTotalPrice() float64 {
	return p.Price + (p.Price * p.TaxRate / 100)
}

// IsInStock checks if product is in stock
func (p *Product) IsInStock() bool {
	return p.Stock > 0
}

// IsLowStock checks if product is low on stock
func (p *Product) IsLowStock() bool {
	return p.Stock <= p.MinStock
}

// generateSKU generates a simple SKU from product name
func generateSKU(name string) string {
	// Simple SKU generation - in production, use a more sophisticated method
	return "SKU-" + name[:min(3, len(name))] + "-" + time.Now().Format("20060102")
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
