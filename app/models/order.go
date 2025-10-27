package models

import (
	"time"
	"gorm.io/gorm"
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

// PaymentStatus represents the payment status
type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// Order represents a sales order in the POS system
type Order struct {
	ID            uint           `json:"id" gorm:"primarykey"`
	OrderNumber   string         `json:"order_number" gorm:"uniqueIndex;not null"`
	CustomerID    *uint          `json:"customer_id"`
	Customer      *Customer      `json:"customer" gorm:"foreignKey:CustomerID"`
	CashierID     uint           `json:"cashier_id"`
	Cashier       User           `json:"cashier" gorm:"foreignKey:CashierID"`
	Status        OrderStatus    `json:"status" gorm:"default:'pending'"`
	PaymentStatus PaymentStatus  `json:"payment_status" gorm:"default:'pending'"`
	Subtotal      float64        `json:"subtotal" gorm:"type:decimal(10,2)"`
	TaxAmount     float64        `json:"tax_amount" gorm:"type:decimal(10,2)"`
	DiscountAmount float64       `json:"discount_amount" gorm:"type:decimal(10,2)"`
	TotalAmount   float64        `json:"total_amount" gorm:"type:decimal(10,2)"`
	Items         []OrderItem    `json:"items" gorm:"foreignKey:OrderID"`
	Payments      []Payment      `json:"payments" gorm:"foreignKey:OrderID"`
	Notes         string         `json:"notes"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the order model
func (Order) TableName() string {
	return "orders"
}

// BeforeCreate is called before creating a new record
func (m *Order) BeforeCreate(tx *gorm.DB) error {
	// Generate order number if not provided
	if m.OrderNumber == "" {
		m.OrderNumber = generateOrderNumber()
	}
	return nil
}

// BeforeUpdate is called before updating a record
func (m *Order) BeforeUpdate(tx *gorm.DB) error {
	// Add any pre-update logic here
	return nil
}

// BeforeDelete is called before deleting a record
func (m *Order) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// CalculateTotals calculates subtotal, tax, and total amounts
func (o *Order) CalculateTotals() {
	o.Subtotal = 0
	o.TaxAmount = 0
	
	for _, item := range o.Items {
		itemTotal := item.Price * float64(item.Quantity)
		o.Subtotal += itemTotal
		o.TaxAmount += itemTotal * item.TaxRate / 100
	}
	
	o.TotalAmount = o.Subtotal + o.TaxAmount - o.DiscountAmount
}

// IsPaid checks if the order is fully paid
func (o *Order) IsPaid() bool {
	return o.PaymentStatus == PaymentStatusPaid
}

// CanBeCancelled checks if the order can be cancelled
func (o *Order) CanBeCancelled() bool {
	return o.Status == OrderStatusPending
}

// generateOrderNumber generates a unique order number
func generateOrderNumber() string {
	return "ORD-" + time.Now().Format("20060102") + "-" + time.Now().Format("150405")
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	OrderID   uint           `json:"order_id"`
	Order     Order          `json:"order" gorm:"foreignKey:OrderID"`
	ProductID uint           `json:"product_id"`
	Product   Product        `json:"product" gorm:"foreignKey:ProductID"`
	Quantity  int            `json:"quantity" gorm:"not null"`
	Price     float64        `json:"price" gorm:"type:decimal(10,2)"`
	TaxRate   float64        `json:"tax_rate" gorm:"type:decimal(5,2)"`
	Discount  float64        `json:"discount" gorm:"type:decimal(10,2)"`
	Total     float64        `json:"total" gorm:"type:decimal(10,2)"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName returns the table name for the order item model
func (OrderItem) TableName() string {
	return "order_items"
}

// BeforeCreate is called before creating a new record
func (m *OrderItem) BeforeCreate(tx *gorm.DB) error {
	// Calculate total
	m.Total = (m.Price * float64(m.Quantity)) - m.Discount
	return nil
}

// BeforeUpdate is called before updating a record
func (m *OrderItem) BeforeUpdate(tx *gorm.DB) error {
	// Recalculate total
	m.Total = (m.Price * float64(m.Quantity)) - m.Discount
	return nil
}

// BeforeDelete is called before deleting a record
func (m *OrderItem) BeforeDelete(tx *gorm.DB) error {
	// Add any pre-delete logic here
	return nil
}

// GetLineTotal calculates the line total for this item
func (oi *OrderItem) GetLineTotal() float64 {
	return (oi.Price * float64(oi.Quantity)) - oi.Discount
}
