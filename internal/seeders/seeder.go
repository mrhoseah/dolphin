package seeders

import (
	"fmt"
	"log"

	"github.com/mrhoseah/dolphin/internal/factories"
	"gorm.io/gorm"
)

// Seeder defines the interface for database seeders
type Seeder interface {
	// Run executes the seeder
	Run(db *gorm.DB) error
	
	// GetName returns the seeder name
	GetName() string
	
	// GetDependencies returns seeder dependencies
	GetDependencies() []string
}

// BaseSeeder provides common seeder functionality
type BaseSeeder struct {
	name         string
	dependencies []string
	factoryManager *factories.FactoryManager
}

// NewBaseSeeder creates a new base seeder
func NewBaseSeeder(name string, dependencies ...string) *BaseSeeder {
	return &BaseSeeder{
		name:         name,
		dependencies: dependencies,
	}
}

// GetName returns the seeder name
func (bs *BaseSeeder) GetName() string {
	return bs.name
}

// GetDependencies returns seeder dependencies
func (bs *BaseSeeder) GetDependencies() []string {
	return bs.dependencies
}

// SetFactoryManager sets the factory manager
func (bs *BaseSeeder) SetFactoryManager(fm *factories.FactoryManager) {
	bs.factoryManager = fm
}

// SeederManager manages all seeders
type SeederManager struct {
	seeders map[string]Seeder
	db      *gorm.DB
	factoryManager *factories.FactoryManager
}

// NewSeederManager creates a new seeder manager
func NewSeederManager(db *gorm.DB) *SeederManager {
	return &SeederManager{
		seeders: make(map[string]Seeder),
		db:      db,
		factoryManager: factories.NewFactoryManager(db),
	}
}

// Register registers a seeder
func (sm *SeederManager) Register(seeder Seeder) {
	sm.seeders[seeder.GetName()] = seeder
	
	// Set factory manager if it's a BaseSeeder
	if baseSeeder, ok := seeder.(*BaseSeeder); ok {
		baseSeeder.SetFactoryManager(sm.factoryManager)
	}
}

// Run runs a specific seeder
func (sm *SeederManager) Run(name string) error {
	seeder, exists := sm.seeders[name]
	if !exists {
		return fmt.Errorf("seeder '%s' not found", name)
	}
	
	// Check if already run
	if sm.isSeederRun(name) {
		log.Printf("Seeder '%s' already run, skipping", name)
		return nil
	}
	
	// Run dependencies first
	for _, dep := range seeder.GetDependencies() {
		if err := sm.Run(dep); err != nil {
			return fmt.Errorf("failed to run dependency '%s': %w", dep, err)
		}
	}
	
	// Run the seeder
	log.Printf("Running seeder: %s", name)
	if err := seeder.Run(sm.db); err != nil {
		return fmt.Errorf("failed to run seeder '%s': %w", name, err)
	}
	
	// Mark as run
	if err := sm.markSeederRun(name); err != nil {
		log.Printf("Warning: failed to mark seeder '%s' as run: %v", name, err)
	}
	
	log.Printf("Seeder '%s' completed successfully", name)
	return nil
}

// RunAll runs all registered seeders
func (sm *SeederManager) RunAll() error {
	// Create a dependency graph and run in order
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)
	
	var runOrder []string
	
	for name := range sm.seeders {
		if !visited[name] {
			if err := sm.topologicalSort(name, visited, recursionStack, &runOrder); err != nil {
				return err
			}
		}
	}
	
	// Run seeders in topological order
	for _, name := range runOrder {
		if err := sm.Run(name); err != nil {
			return err
		}
	}
	
	return nil
}

// topologicalSort performs topological sort for dependency resolution
func (sm *SeederManager) topologicalSort(name string, visited, recursionStack map[string]bool, result *[]string) error {
	visited[name] = true
	recursionStack[name] = true
	
	seeder := sm.seeders[name]
	for _, dep := range seeder.GetDependencies() {
		if !visited[dep] {
			if err := sm.topologicalSort(dep, visited, recursionStack, result); err != nil {
				return err
			}
		} else if recursionStack[dep] {
			return fmt.Errorf("circular dependency detected: %s -> %s", name, dep)
		}
	}
	
	recursionStack[name] = false
	*result = append(*result, name)
	return nil
}

// isSeederRun checks if a seeder has been run
func (sm *SeederManager) isSeederRun(name string) bool {
	var count int64
	sm.db.Table("seeders").Where("name = ?", name).Count(&count)
	return count > 0
}

// markSeederRun marks a seeder as run
func (sm *SeederManager) markSeederRun(name string) error {
	return sm.db.Table("seeders").Create(map[string]interface{}{
		"name": name,
	}).Error
}

// UserSeeder seeds user data
type UserSeeder struct {
	*BaseSeeder
}

// NewUserSeeder creates a new user seeder
func NewUserSeeder() *UserSeeder {
	return &UserSeeder{
		BaseSeeder: NewBaseSeeder("users"),
	}
}

// Run executes the user seeder
func (us *UserSeeder) Run(db *gorm.DB) error {
	// Create users table if it doesn't exist
	if err := db.AutoMigrate(&factories.User{}); err != nil {
		return err
	}
	
	// Create seeders table if it doesn't exist
	if err := db.Exec("CREATE TABLE IF NOT EXISTS seeders (name VARCHAR(255) PRIMARY KEY, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP)").Error; err != nil {
		return err
	}
	
	// Register user factory
	userFactory := factories.NewUserFactory(db)
	us.factoryManager.Register("users", userFactory)
	
	// Create admin user
	admin := userFactory.Admin().Create(map[string]interface{}{
		"Name":     "Admin User",
		"Email":    "admin@example.com",
		"Username": "admin",
		"Password": "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password
	}).(*factories.User)
	
	log.Printf("Created admin user: %s (%s)", admin.Name, admin.Email)
	
	// Create regular users
	users := userFactory.CreateMany(50)
	log.Printf("Created %d regular users", len(users))
	
	// Create inactive users
	inactiveUsers := userFactory.Inactive().CreateMany(5)
	log.Printf("Created %d inactive users", len(inactiveUsers))
	
	return nil
}

// PostSeeder seeds post data
type PostSeeder struct {
	*BaseSeeder
}

// NewPostSeeder creates a new post seeder
func NewPostSeeder() *PostSeeder {
	return &PostSeeder{
		BaseSeeder: NewBaseSeeder("posts", "users"), // Depends on users
	}
}

// Run executes the post seeder
func (ps *PostSeeder) Run(db *gorm.DB) error {
	// Create posts table if it doesn't exist
	if err := db.AutoMigrate(&factories.Post{}); err != nil {
		return err
	}
	
	// Register post factory
	postFactory := factories.NewPostFactory(db)
	ps.factoryManager.Register("posts", postFactory)
	
	// Get some users to be authors
	var users []factories.User
	db.Limit(10).Find(&users)
	
	if len(users) == 0 {
		return fmt.Errorf("no users found, run user seeder first")
	}
	
	// Create published posts
	for i := 0; i < 100; i++ {
		author := users[i%len(users)]
		postFactory.Published().WithAuthor(&author).Create()
	}
	log.Printf("Created 100 published posts")
	
	// Create draft posts
	for i := 0; i < 20; i++ {
		author := users[i%len(users)]
		postFactory.Draft().WithAuthor(&author).Create()
	}
	log.Printf("Created 20 draft posts")
	
	// Create featured posts
	for i := 0; i < 10; i++ {
		author := users[i%len(users)]
		postFactory.Featured().WithAuthor(&author).Create()
	}
	log.Printf("Created 10 featured posts")
	
	return nil
}

// ProductSeeder seeds product data
type ProductSeeder struct {
	*BaseSeeder
}

// NewProductSeeder creates a new product seeder
func NewProductSeeder() *ProductSeeder {
	return &ProductSeeder{
		BaseSeeder: NewBaseSeeder("products"),
	}
}

// Run executes the product seeder
func (ps *ProductSeeder) Run(db *gorm.DB) error {
	// Create products table if it doesn't exist
	if err := db.AutoMigrate(&factories.Product{}); err != nil {
		return err
	}
	
	// Register product factory
	productFactory := factories.NewProductFactory(db)
	ps.factoryManager.Register("products", productFactory)
	
	// Create regular products
	products := productFactory.CreateMany(200)
	log.Printf("Created %d regular products", len(products))
	
	// Create featured products
	featuredProducts := productFactory.Featured().CreateMany(20)
	log.Printf("Created %d featured products", len(featuredProducts))
	
	// Create out of stock products
	outOfStockProducts := productFactory.OutOfStock().CreateMany(10)
	log.Printf("Created %d out of stock products", len(outOfStockProducts))
	
	// Create high-rated products
	highRatedProducts := productFactory.HighRated().CreateMany(30)
	log.Printf("Created %d high-rated products", len(highRatedProducts))
	
	return nil
}

// OrderSeeder seeds order data
type OrderSeeder struct {
	*BaseSeeder
}

// NewOrderSeeder creates a new order seeder
func NewOrderSeeder() *OrderSeeder {
	return &OrderSeeder{
		BaseSeeder: NewBaseSeeder("orders", "users"), // Depends on users
	}
}

// Run executes the order seeder
func (os *OrderSeeder) Run(db *gorm.DB) error {
	// Create orders table if it doesn't exist
	if err := db.AutoMigrate(&factories.Order{}); err != nil {
		return err
	}
	
	// Register order factory
	orderFactory := factories.NewOrderFactory(db)
	os.factoryManager.Register("orders", orderFactory)
	
	// Get some users to be customers
	var users []factories.User
	db.Limit(20).Find(&users)
	
	if len(users) == 0 {
		return fmt.Errorf("no users found, run user seeder first")
	}
	
	// Create pending orders
	for i := 0; i < 50; i++ {
		customer := users[i%len(users)]
		orderFactory.WithCustomer(&customer).Create()
	}
	log.Printf("Created 50 pending orders")
	
	// Create completed orders
	for i := 0; i < 100; i++ {
		customer := users[i%len(users)]
		orderFactory.Completed().WithCustomer(&customer).Create()
	}
	log.Printf("Created 100 completed orders")
	
	// Create cancelled orders
	for i := 0; i < 10; i++ {
		customer := users[i%len(users)]
		orderFactory.Cancelled().WithCustomer(&customer).Create()
	}
	log.Printf("Created 10 cancelled orders")
	
	return nil
}

// DatabaseSeeder seeds all data
type DatabaseSeeder struct {
	*BaseSeeder
}

// NewDatabaseSeeder creates a new database seeder
func NewDatabaseSeeder() *DatabaseSeeder {
	return &DatabaseSeeder{
		BaseSeeder: NewBaseSeeder("database", "users", "posts", "products", "orders"),
	}
}

// Run executes the database seeder
func (ds *DatabaseSeeder) Run(db *gorm.DB) error {
	log.Println("Running database seeder...")
	
	// This seeder doesn't create data itself, it just ensures all other seeders run
	// The dependencies will be handled by the SeederManager
	
	log.Println("Database seeding completed!")
	return nil
}
