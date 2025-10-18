package factories

import (
	"time"

	"gorm.io/gorm"
)

// UserFactory creates User model instances
type UserFactory struct {
	*BaseFactory
	fake *FakeData
}

// NewUserFactory creates a new user factory
func NewUserFactory(db *gorm.DB) *UserFactory {
	user := &User{} // Assuming User model exists
	baseFactory := NewFactory(user, db)

	return &UserFactory{
		BaseFactory: baseFactory,
		fake:        NewFakeData(),
	}
}

// User represents a user model (example)
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email" gorm:"uniqueIndex"`
	Username  string         `json:"username" gorm:"uniqueIndex"`
	Password  string         `json:"-"`
	Phone     string         `json:"phone"`
	Address   string         `json:"address"`
	City      string         `json:"city"`
	Country   string         `json:"country"`
	Company   string         `json:"company"`
	JobTitle  string         `json:"job_title"`
	Bio       string         `json:"bio"`
	Avatar    string         `json:"avatar"`
	IsActive  bool           `json:"is_active"`
	IsAdmin   bool           `json:"is_admin"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// getDefaultAttributes returns default user attributes
func (uf *UserFactory) getDefaultAttributes() map[string]interface{} {
	return map[string]interface{}{
		"Name":     uf.fake.Name(),
		"Email":    uf.fake.Email(),
		"Username": uf.fake.Username(),
		"Password": uf.fake.Password(),
		"Phone":    uf.fake.Phone(),
		"Address":  uf.fake.Address(),
		"City":     uf.fake.City(),
		"Country":  uf.fake.Country(),
		"Company":  uf.fake.Company(),
		"JobTitle": uf.fake.JobTitle(),
		"Bio":      uf.fake.Paragraph(),
		"Avatar":   uf.fake.ImageURL(),
		"IsActive": true,
		"IsAdmin":  false,
	}
}

// Admin creates an admin user
func (uf *UserFactory) Admin() *UserFactory {
	uf.State("admin", func(instance interface{}) {
		if user, ok := instance.(*User); ok {
			user.IsAdmin = true
			user.IsActive = true
		}
	})
	return uf
}

// Inactive creates an inactive user
func (uf *UserFactory) Inactive() *UserFactory {
	uf.State("inactive", func(instance interface{}) {
		if user, ok := instance.(*User); ok {
			user.IsActive = false
		}
	})
	return uf
}

// Verified creates a verified user
func (uf *UserFactory) Verified() *UserFactory {
	uf.State("verified", func(instance interface{}) {
		if user, ok := instance.(*User); ok {
			user.IsActive = true
		}
	})
	return uf
}

// PostFactory creates Post model instances
type PostFactory struct {
	*BaseFactory
	fake *FakeData
}

// NewPostFactory creates a new post factory
func NewPostFactory(db *gorm.DB) *PostFactory {
	post := &Post{}
	baseFactory := NewFactory(post, db)

	return &PostFactory{
		BaseFactory: baseFactory,
		fake:        NewFakeData(),
	}
}

// Post represents a post model (example)
type Post struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `json:"title"`
	Content     string         `json:"content"`
	Excerpt     string         `json:"excerpt"`
	Slug        string         `json:"slug" gorm:"uniqueIndex"`
	AuthorID    uint           `json:"author_id"`
	Author      *User          `json:"author" gorm:"foreignKey:AuthorID"`
	Status      string         `json:"status"` // draft, published, archived
	Views       int            `json:"views"`
	Likes       int            `json:"likes"`
	Featured    bool           `json:"featured"`
	PublishedAt *time.Time     `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// getDefaultAttributes returns default post attributes
func (pf *PostFactory) getDefaultAttributes() map[string]interface{} {
	title := pf.fake.Sentence()
	return map[string]interface{}{
		"Title":       title,
		"Content":     pf.fake.Paragraph(),
		"Excerpt":     pf.fake.Sentence(),
		"Slug":        pf.fake.Username(), // Simplified slug generation
		"Status":      "published",
		"Views":       pf.fake.Number(0, 1000),
		"Likes":       pf.fake.Number(0, 100),
		"Featured":    false,
		"PublishedAt": time.Now(),
	}
}

// Draft creates a draft post
func (pf *PostFactory) Draft() *PostFactory {
	pf.State("draft", func(instance interface{}) {
		if post, ok := instance.(*Post); ok {
			post.Status = "draft"
			post.PublishedAt = nil
		}
	})
	return pf
}

// Published creates a published post
func (pf *PostFactory) Published() *PostFactory {
	pf.State("published", func(instance interface{}) {
		if post, ok := instance.(*Post); ok {
			post.Status = "published"
			now := time.Now()
			post.PublishedAt = &now
		}
	})
	return pf
}

// Featured creates a featured post
func (pf *PostFactory) Featured() *PostFactory {
	pf.State("featured", func(instance interface{}) {
		if post, ok := instance.(*Post); ok {
			post.Featured = true
			post.Status = "published"
			now := time.Now()
			post.PublishedAt = &now
		}
	})
	return pf
}

// WithAuthor creates a post with a specific author
func (pf *PostFactory) WithAuthor(author *User) *PostFactory {
	pf.State("with_author", func(instance interface{}) {
		if post, ok := instance.(*Post); ok {
			post.AuthorID = author.ID
			post.Author = author
		}
	})
	return pf
}

// ProductFactory creates Product model instances
type ProductFactory struct {
	*BaseFactory
	fake *FakeData
}

// NewProductFactory creates a new product factory
func NewProductFactory(db *gorm.DB) *ProductFactory {
	product := &Product{}
	baseFactory := NewFactory(product, db)

	return &ProductFactory{
		BaseFactory: baseFactory,
		fake:        NewFakeData(),
	}
}

// Product represents a product model (example)
type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	SKU         string         `json:"sku" gorm:"uniqueIndex"`
	Price       float64        `json:"price"`
	Cost        float64        `json:"cost"`
	Stock       int            `json:"stock"`
	Category    string         `json:"category"`
	Brand       string         `json:"brand"`
	Image       string         `json:"image"`
	Weight      float64        `json:"weight"`
	Dimensions  string         `json:"dimensions"`
	IsActive    bool           `json:"is_active"`
	IsFeatured  bool           `json:"is_featured"`
	Rating      float64        `json:"rating"`
	Reviews     int            `json:"reviews"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// getDefaultAttributes returns default product attributes
func (pf *ProductFactory) getDefaultAttributes() map[string]interface{} {
	return map[string]interface{}{
		"Name":        pf.fake.Sentence(),
		"Description": pf.fake.Paragraph(),
		"SKU":         pf.fake.UUID(),
		"Price":       pf.fake.Price(10.0, 1000.0),
		"Cost":        pf.fake.Price(5.0, 500.0),
		"Stock":       pf.fake.Number(0, 1000),
		"Category":    pf.fake.RandomElement([]string{"Electronics", "Clothing", "Books", "Home", "Sports"}),
		"Brand":       pf.fake.Company(),
		"Image":       pf.fake.ImageURL(),
		"Weight":      pf.fake.Float(0.1, 50.0),
		"Dimensions":  "10x10x5 cm",
		"IsActive":    true,
		"IsFeatured":  false,
		"Rating":      pf.fake.Float(1.0, 5.0),
		"Reviews":     pf.fake.Number(0, 500),
	}
}

// Featured creates a featured product
func (pf *ProductFactory) Featured() *ProductFactory {
	pf.State("featured", func(instance interface{}) {
		if product, ok := instance.(*Product); ok {
			product.IsFeatured = true
			product.IsActive = true
		}
	})
	return pf
}

// OutOfStock creates an out of stock product
func (pf *ProductFactory) OutOfStock() *ProductFactory {
	pf.State("out_of_stock", func(instance interface{}) {
		if product, ok := instance.(*Product); ok {
			product.Stock = 0
			product.IsActive = false
		}
	})
	return pf
}

// HighRated creates a high-rated product
func (pf *ProductFactory) HighRated() *ProductFactory {
	pf.State("high_rated", func(instance interface{}) {
		if product, ok := instance.(*Product); ok {
			product.Rating = pf.fake.Float(4.0, 5.0)
			product.Reviews = pf.fake.Number(100, 1000)
		}
	})
	return pf
}

// OrderFactory creates Order model instances
type OrderFactory struct {
	*BaseFactory
	fake *FakeData
}

// NewOrderFactory creates a new order factory
func NewOrderFactory(db *gorm.DB) *OrderFactory {
	order := &Order{}
	baseFactory := NewFactory(order, db)

	return &OrderFactory{
		BaseFactory: baseFactory,
		fake:        NewFakeData(),
	}
}

// Order represents an order model (example)
type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	OrderNumber     string         `json:"order_number" gorm:"uniqueIndex"`
	CustomerID      uint           `json:"customer_id"`
	Customer        *User          `json:"customer" gorm:"foreignKey:CustomerID"`
	Status          string         `json:"status"` // pending, processing, shipped, delivered, cancelled
	Total           float64        `json:"total"`
	Subtotal        float64        `json:"subtotal"`
	Tax             float64        `json:"tax"`
	Shipping        float64        `json:"shipping"`
	Discount        float64        `json:"discount"`
	Currency        string         `json:"currency"`
	PaymentMethod   string         `json:"payment_method"`
	ShippingAddress string         `json:"shipping_address"`
	BillingAddress  string         `json:"billing_address"`
	Notes           string         `json:"notes"`
	ShippedAt       *time.Time     `json:"shipped_at"`
	DeliveredAt     *time.Time     `json:"delivered_at"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// getDefaultAttributes returns default order attributes
func (of *OrderFactory) getDefaultAttributes() map[string]interface{} {
	subtotal := of.fake.Price(50.0, 500.0)
	tax := subtotal * 0.1
	shipping := of.fake.Price(5.0, 25.0)
	total := subtotal + tax + shipping

	return map[string]interface{}{
		"OrderNumber":     of.fake.UUID(),
		"Status":          "pending",
		"Total":           total,
		"Subtotal":        subtotal,
		"Tax":             tax,
		"Shipping":        shipping,
		"Discount":        0.0,
		"Currency":        "USD",
		"PaymentMethod":   of.fake.RandomElement([]string{"credit_card", "paypal", "bank_transfer"}),
		"ShippingAddress": of.fake.Address(),
		"BillingAddress":  of.fake.Address(),
		"Notes":           of.fake.Sentence(),
	}
}

// Completed creates a completed order
func (of *OrderFactory) Completed() *OrderFactory {
	of.State("completed", func(instance interface{}) {
		if order, ok := instance.(*Order); ok {
			order.Status = "delivered"
			now := time.Now()
			order.DeliveredAt = &now
			shippedAt := now.Add(-24 * time.Hour)
			order.ShippedAt = &shippedAt
		}
	})
	return of
}

// Cancelled creates a cancelled order
func (of *OrderFactory) Cancelled() *OrderFactory {
	of.State("cancelled", func(instance interface{}) {
		if order, ok := instance.(*Order); ok {
			order.Status = "cancelled"
		}
	})
	return of
}

// WithCustomer creates an order with a specific customer
func (of *OrderFactory) WithCustomer(customer *User) *OrderFactory {
	of.State("with_customer", func(instance interface{}) {
		if order, ok := instance.(*Order); ok {
			order.CustomerID = customer.ID
			order.Customer = customer
		}
	})
	return of
}
