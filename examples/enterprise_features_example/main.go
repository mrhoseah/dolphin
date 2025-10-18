package main

import (
	"fmt"
	"log"
	"time"

	"dolphin/internal/factories"
	"dolphin/internal/fakes"
	"dolphin/internal/httpclient"
	"dolphin/internal/seeders"
	"dolphin/internal/validation"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	fmt.Println("🚀 Dolphin Framework - Enterprise Features Demo")
	fmt.Println("================================================")
	fmt.Println("")

	// 1. Database Factories and Seeders Demo
	fmt.Println("📊 1. Database Factories and Seeders")
	fmt.Println("-----------------------------------")
	demoFactoriesAndSeeders()

	fmt.Println("")

	// 2. Mocking/Fakes System Demo
	fmt.Println("🎭 2. Mocking/Fakes System")
	fmt.Println("--------------------------")
	demoMockingSystem()

	fmt.Println("")

	// 3. First-Party HTTP Client Demo
	fmt.Println("🌐 3. First-Party HTTP Client")
	fmt.Println("-----------------------------")
	demoHTTPClient()

	fmt.Println("")

	// 4. Standardized Validation Layer Demo
	fmt.Println("✅ 4. Standardized Validation Layer")
	fmt.Println("----------------------------------")
	demoValidationLayer()

	fmt.Println("")
	fmt.Println("🎉 All enterprise features demonstrated successfully!")
}

func demoFactoriesAndSeeders() {
	// Setup database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&factories.User{}, &factories.Post{}, &factories.Product{}, &factories.Order{})

	// Create factory manager
	factoryManager := factories.NewFactoryManager(db)

	// Register factories
	userFactory := factories.NewUserFactory(db)
	postFactory := factories.NewPostFactory(db)
	productFactory := factories.NewProductFactory(db)
	orderFactory := factories.NewOrderFactory(db)

	factoryManager.Register("users", userFactory)
	factoryManager.Register("posts", postFactory)
	factoryManager.Register("products", productFactory)
	factoryManager.Register("orders", orderFactory)

	// Create users using factories
	fmt.Println("Creating users with factories...")

	// Create admin user
	admin := userFactory.Admin().Create(map[string]interface{}{
		"Name":     "Admin User",
		"Email":    "admin@example.com",
		"Username": "admin",
	}).(*factories.User)
	fmt.Printf("✅ Created admin user: %s (%s)\n", admin.Name, admin.Email)

	// Create regular users
	users := userFactory.CreateMany(5)
	fmt.Printf("✅ Created %d regular users\n", len(users))

	// Create inactive users
	inactiveUsers := userFactory.Inactive().CreateMany(2)
	fmt.Printf("✅ Created %d inactive users\n", len(inactiveUsers))

	// Create posts with relationships
	fmt.Println("\nCreating posts with relationships...")
	var allUsers []factories.User
	db.Find(&allUsers)

	for i := 0; i < 10; i++ {
		author := allUsers[i%len(allUsers)]
		post := postFactory.Published().WithAuthor(&author).Create().(*factories.Post)
		fmt.Printf("✅ Created post: '%s' by %s\n", post.Title, author.Name)
	}

	// Create featured posts
	for i := 0; i < 3; i++ {
		author := allUsers[i%len(allUsers)]
		post := postFactory.Featured().WithAuthor(&author).Create().(*factories.Post)
		fmt.Printf("✅ Created featured post: '%s' by %s\n", post.Title, author.Name)
	}

	// Create products
	fmt.Println("\nCreating products...")
	products := productFactory.CreateMany(20)
	fmt.Printf("✅ Created %d regular products\n", len(products))

	featuredProducts := productFactory.Featured().CreateMany(5)
	fmt.Printf("✅ Created %d featured products\n", len(featuredProducts))

	// Create orders with relationships
	fmt.Println("\nCreating orders with relationships...")
	for i := 0; i < 15; i++ {
		customer := allUsers[i%len(allUsers)]
		order := orderFactory.WithCustomer(&customer).Create().(*factories.Order)
		fmt.Printf("✅ Created order: %s for %s (Total: $%.2f)\n",
			order.OrderNumber, customer.Name, order.Total)
	}

	// Demo seeders
	fmt.Println("\nRunning database seeders...")
	seederManager := seeders.NewSeederManager(db)

	// Register seeders
	seederManager.Register(seeders.NewUserSeeder())
	seederManager.Register(seeders.NewPostSeeder())
	seederManager.Register(seeders.NewProductSeeder())
	seederManager.Register(seeders.NewOrderSeeder())
	seederManager.Register(seeders.NewDatabaseSeeder())

	// Run specific seeder
	if err := seederManager.Run("users"); err != nil {
		fmt.Printf("❌ Seeder error: %v\n", err)
	} else {
		fmt.Println("✅ User seeder completed")
	}

	// Count final records
	var userCount, postCount, productCount, orderCount int64
	db.Model(&factories.User{}).Count(&userCount)
	db.Model(&factories.Post{}).Count(&postCount)
	db.Model(&factories.Product{}).Count(&productCount)
	db.Model(&factories.Order{}).Count(&orderCount)

	fmt.Printf("\n📊 Final Database Counts:\n")
	fmt.Printf("  Users: %d\n", userCount)
	fmt.Printf("  Posts: %d\n", postCount)
	fmt.Printf("  Products: %d\n", productCount)
	fmt.Printf("  Orders: %d\n", orderCount)
}

func demoMockingSystem() {
	// Create fake manager
	fakeManager := fakes.NewFakeManager()

	// Create fake services
	fakeCache := fakes.NewFakeCache()
	fakeStorage := fakes.NewFakeFileStorage()
	fakeEvents := fakes.NewFakeEventDispatcher()
	fakeMailer := fakes.NewFakeMailer()
	fakeQueue := fakes.NewFakeQueue()

	// Register fake services
	fakeManager.Register("cache", fakeCache)
	fakeManager.Register("storage", fakeStorage)
	fakeManager.Register("events", fakeEvents)
	fakeManager.Register("mailer", fakeMailer)
	fakeManager.Register("queue", fakeQueue)

	// Fake all services
	fakeManager.FakeAll()
	fmt.Println("✅ All services faked")

	// Demo fake cache
	fmt.Println("\nTesting fake cache...")
	fakeCache.Set("user:1", "John Doe", 5*time.Minute)
	fakeCache.Set("session:abc123", "active", 30*time.Minute)

	if value, exists := fakeCache.Get("user:1"); exists {
		fmt.Printf("✅ Cache hit: user:1 = %s\n", value)
	}

	if fakeCache.Has("session:abc123") {
		fmt.Println("✅ Session exists in cache")
	}

	// Demo fake file storage
	fmt.Println("\nTesting fake file storage...")
	fakeStorage.Put("uploads/avatar.jpg", []byte("fake image data"))
	fakeStorage.Put("documents/report.pdf", []byte("fake pdf data"))

	if fakeStorage.Exists("uploads/avatar.jpg") {
		fmt.Println("✅ Avatar file exists in fake storage")
	}

	files := fakeStorage.List("uploads/")
	fmt.Printf("✅ Found %d files in uploads directory\n", len(files))

	// Demo fake event dispatcher
	fmt.Println("\nTesting fake event dispatcher...")
	fakeEvents.Dispatch("user.registered", map[string]interface{}{
		"user_id": 1,
		"email":   "user@example.com",
	})
	fakeEvents.Dispatch("order.created", map[string]interface{}{
		"order_id": 123,
		"total":    99.99,
	})

	events := fakeEvents.GetEvents()
	fmt.Printf("✅ Dispatched %d events\n", len(events))

	userEvents := fakeEvents.GetEventsByName("user.registered")
	fmt.Printf("✅ Found %d user.registered events\n", len(userEvents))

	// Demo fake mailer
	fmt.Println("\nTesting fake mailer...")
	fakeMailer.Send([]string{"user@example.com"}, "Welcome!", "Welcome to our platform!", "<h1>Welcome!</h1>")
	fakeMailer.Send([]string{"admin@example.com"}, "New User", "A new user has registered", "<p>New user registered</p>")

	sentMails := fakeMailer.GetSentMails()
	fmt.Printf("✅ Sent %d emails\n", len(sentMails))

	for _, mail := range sentMails {
		fmt.Printf("  📧 To: %s, Subject: %s\n", mail.To[0], mail.Subject)
	}

	// Demo fake queue
	fmt.Println("\nTesting fake queue...")
	fakeQueue.Push("send-email", map[string]interface{}{
		"to":      "user@example.com",
		"subject": "Welcome!",
	})
	fakeQueue.Push("process-payment", map[string]interface{}{
		"order_id": 123,
		"amount":   99.99,
	})

	jobs := fakeQueue.GetJobs()
	fmt.Printf("✅ Queued %d jobs\n", len(jobs))

	emailJobs := fakeQueue.GetJobsByName("send-email")
	fmt.Printf("✅ Found %d email jobs\n", len(emailJobs))

	// Show faked services
	fakedServices := fakeManager.GetFakedServices()
	fmt.Printf("\n🎭 Currently faked services: %v\n", fakedServices)

	// Restore all services
	fakeManager.RestoreAll()
	fmt.Println("✅ All services restored to real implementations")
}

func demoHTTPClient() {
	// Create HTTP client
	config := &httpclient.Config{
		BaseURL: "https://jsonplaceholder.typicode.com",
		Timeout: 10 * time.Second,
		Retries: 3,
		Headers: map[string]string{
			"User-Agent": "Dolphin-Framework-Demo/1.0.0",
		},
	}

	client := httpclient.NewClient(config)
	fmt.Println("✅ HTTP client created with configuration")

	// Demo GET request
	fmt.Println("\nTesting GET request...")
	resp, err := client.Get("/posts/1")
	if err != nil {
		fmt.Printf("❌ GET request failed: %v\n", err)
		return
	}

	if resp.IsSuccess() {
		fmt.Printf("✅ GET request successful (Status: %d)\n", resp.StatusCode)

		var post map[string]interface{}
		if err := resp.JSON(&post); err == nil {
			fmt.Printf("  📄 Post Title: %s\n", post["title"])
			fmt.Printf("  📄 Post ID: %.0f\n", post["id"])
		}
	}

	// Demo POST request with fluent builder
	fmt.Println("\nTesting POST request with fluent builder...")
	newPost := map[string]interface{}{
		"title":  "Dolphin Framework Demo Post",
		"body":   "This is a test post created by Dolphin Framework",
		"userId": 1,
	}

	resp, err = client.POST("/posts").
		Body(newPost).
		Header("Content-Type", "application/json").
		Send()

	if err != nil {
		fmt.Printf("❌ POST request failed: %v\n", err)
		return
	}

	if resp.IsSuccess() {
		fmt.Printf("✅ POST request successful (Status: %d)\n", resp.StatusCode)

		var createdPost map[string]interface{}
		if err := resp.JSON(&createdPost); err == nil {
			fmt.Printf("  📄 Created Post ID: %.0f\n", createdPost["id"])
		}
	}

	// Demo request with query parameters
	fmt.Println("\nTesting request with query parameters...")
	resp, err = client.GET("/posts").
		QueryParam("userId", "1").
		QueryParam("_limit", "3").
		Send()

	if err != nil {
		fmt.Printf("❌ Query request failed: %v\n", err)
		return
	}

	if resp.IsSuccess() {
		fmt.Printf("✅ Query request successful (Status: %d)\n", resp.StatusCode)

		var posts []map[string]interface{}
		if err := resp.JSON(&posts); err == nil {
			fmt.Printf("  📄 Found %d posts for user 1\n", len(posts))
		}
	}

	// Demo error handling
	fmt.Println("\nTesting error handling...")
	resp, err = client.Get("/nonexistent")
	if err != nil {
		fmt.Printf("❌ Expected error: %v\n", err)
	} else if resp.IsClientError() {
		fmt.Printf("✅ Expected client error (Status: %d)\n", resp.StatusCode)
	}
}

func demoValidationLayer() {
	// Demo basic validation
	fmt.Println("Testing basic validation...")

	request := validation.NewRequest(map[string]interface{}{
		"name":     "John Doe",
		"email":    "john@example.com",
		"age":      25,
		"password": "secret123",
		"website":  "https://johndoe.com",
	})

	// Add validation rules
	request.Required("name").
		StringWithLength("name", 2, 50).
		Required("email").
		Email("email").
		Required("age").
		IntegerRange("age", 18, 100).
		Required("password").
		StringWithLength("password", 8, 128).
		URL("website")

	// Validate
	if request.Validate() {
		fmt.Println("✅ Validation passed")
		fmt.Printf("  Name: %s\n", request.GetString("name"))
		fmt.Printf("  Email: %s\n", request.GetString("email"))
		fmt.Printf("  Age: %d\n", request.GetInt("age"))
	} else {
		fmt.Println("❌ Validation failed")
		for field, errors := range request.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Demo validation with errors
	fmt.Println("\nTesting validation with errors...")

	invalidRequest := validation.NewRequest(map[string]interface{}{
		"name":     "",              // Empty name
		"email":    "invalid-email", // Invalid email
		"age":      15,              // Too young
		"password": "123",           // Too short
		"website":  "not-a-url",     // Invalid URL
	})

	invalidRequest.Required("name").
		StringWithLength("name", 2, 50).
		Required("email").
		Email("email").
		Required("age").
		IntegerRange("age", 18, 100).
		Required("password").
		StringWithLength("password", 8, 128).
		URL("website")

	if !invalidRequest.Validate() {
		fmt.Println("✅ Validation correctly failed")
		for field, errors := range invalidRequest.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}

	// Demo custom validation rule
	fmt.Println("\nTesting custom validation rule...")

	customRequest := validation.NewRequest(map[string]interface{}{
		"username": "admin",
		"role":     "user",
	})

	// Custom rule: username cannot be "admin"
	customRule := &validation.BaseRule{
		Field:   "username",
		Message: "Username 'admin' is reserved",
	}

	customRequest.Custom("username", &validation.RequiredRule{BaseRule: customRule})
	customRequest.Custom("username", &validation.StringRule{BaseRule: customRule})

	// Add custom validation logic
	customRequest.Custom("username", &validation.BaseRule{
		Field:   "username",
		Message: "Username 'admin' is reserved",
	})

	if !customRequest.Validate() {
		fmt.Println("✅ Custom validation correctly failed")
		fmt.Printf("  Username error: %s\n", customRequest.GetError("username"))
	}

	// Demo struct-based validation
	fmt.Println("\nTesting struct-based validation...")

	type UserRegistration struct {
		Name     string `validate:"required|string|min_length:2|max_length:50"`
		Email    string `validate:"required|email"`
		Age      int    `validate:"required|integer|min:18|max:100"`
		Password string `validate:"required|string|min_length:8"`
		Website  string `validate:"url"`
	}

	userData := UserRegistration{
		Name:     "Jane Smith",
		Email:    "jane@example.com",
		Age:      28,
		Password: "securepassword123",
		Website:  "https://janesmith.com",
	}

	// Parse rules from struct
	rules := validation.ParseRulesFromStruct(userData)
	fmt.Printf("✅ Parsed %d validation rules from struct\n", len(rules))

	// Create validator and add rules
	validator := validation.NewValidator()
	for field, fieldRules := range rules {
		validator.AddRules(field, fieldRules...)
	}

	// Convert struct to map for validation
	data := map[string]interface{}{
		"name":     userData.Name,
		"email":    userData.Email,
		"age":      userData.Age,
		"password": userData.Password,
		"website":  userData.Website,
	}

	if validator.Validate(data) {
		fmt.Println("✅ Struct validation passed")
	} else {
		fmt.Println("❌ Struct validation failed")
		for field, errors := range validator.GetErrors() {
			fmt.Printf("  %s: %v\n", field, errors)
		}
	}
}
