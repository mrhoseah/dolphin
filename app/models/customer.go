package models

import (
	"time"
	"gorm.io/gorm"
)

// Customer represents a customer in the POS system
type Customer struct {
	ID           uint           `json:"id" gorm:"primarykey"`
	FirstName    string         `json:"first_name" gorm:"not null"`
	LastName     string         `json:"last_name" gorm:"not null"`
	Email        string         `json:"email" gorm:"uniqueIndex"`
	Phone        string         `json:"phone"`
	Address      string         `json:"address"`
	City         string         `json:"city"`
	State        string         `json:"state"`
	ZipCode      string         `json:"zip_code"`
	Country      string         `json:"country"`
	DateOfBirth  *time.Time     `json:"date_of_birth"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LoyaltyPoints int           `json:"loyalty_points" gorm:"default:0"`
	Orders       []Order        `json:"orders" gorm:"foreignKey:CustomerID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the customer model
func (Customer) TableName() string {
	return "customers"
}

// BeforeCreate is called before creating a new record
func (m *Customer) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	return nil
}

// BeforeUpdate is called before updating a record
func (m *Customer) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *Customer) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// GetFullName returns the customer's full name
func (c *Customer) GetFullName() string {
	return c.FirstName + " " + c.LastName
}

// GetTotalOrders returns the number of orders for this customer
func (c *Customer) GetTotalOrders() int {
	return len(c.Orders)
}

// PaymentMethod represents the payment method
type PaymentMethod string

const (
	PaymentMethodCash      PaymentMethod = "cash"
	PaymentMethodCard      PaymentMethod = "card"
	PaymentMethodMobile    PaymentMethod = "mobile"
	PaymentMethodCheck     PaymentMethod = "check"
	PaymentMethodGiftCard  PaymentMethod = "gift_card"
)

// Payment represents a payment in the POS system
type Payment struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	OrderID       uint           `json:"order_id"`
	Order         Order          `json:"order" gorm:"foreignKey:OrderID"`
	Amount        float64        `json:"amount" gorm:"type:decimal(10,2)"`
	Method        PaymentMethod  `json:"method" gorm:"not null"`
	Reference     string         `json:"reference"`
	Status        PaymentStatus  `json:"status" gorm:"default:'pending'"`
	ProcessedAt   *time.Time     `json:"processed_at"`
	RefundAmount  float64        `json:"refund_amount" gorm:"type:decimal(10,2)"`
	RefundedAt    *time.Time     `json:"refunded_at"`
	Notes         string         `json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the payment model
func (Payment) TableName() string {
	return "payments"
}

// BeforeCreate is called before creating a new record
func (m *Payment) BeforeCreate(tx *gorm.DB) error {
	// Add any pre-create logic here
	return nil
}

// BeforeUpdate is called before updating a record
func (m *Payment) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *Payment) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// IsSuccessful checks if the payment was successful
func (p *Payment) IsSuccessful() bool {
	return p.Status == PaymentStatusPaid
}

// CanBeRefunded checks if the payment can be refunded
func (p *Payment) CanBeRefunded() bool {
	return p.Status == PaymentStatusPaid && p.RefundAmount < p.Amount
}

// GetRefundableAmount returns the amount that can be refunded
func (p *Payment) GetRefundableAmount() float64 {
	return p.Amount - p.RefundAmount
}
