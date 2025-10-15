package app

import (
	"fmt"
	"strings"
)

// generateHTMXViewContent generates HTMX view templates
func (g *Generator) generateHTMXViewContent(name, viewType string) string {
	lowerName := strings.ToLower(name)
	pluralName := lowerName + "s"

	switch viewType {
	case "index":
		return g.generateIndexView(name, lowerName, pluralName)
	case "show":
		return g.generateShowView(name, lowerName)
	case "create":
		return g.generateCreateView(name, lowerName)
	case "edit":
		return g.generateEditView(name, lowerName)
	case "form":
		return g.generateFormPartial(name, lowerName)
	default:
		return ""
	}
}

func (g *Generator) generateIndexView(name, lowerName, pluralName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 %s Management</h1>
                    </div>
                    <div class="flex items-center">
                        <a href="/%s/create" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
                            Create New
                        </a>
                    </div>
                </div>
            </div>
        </nav>
        
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div id="%s-list" class="bg-white rounded-lg shadow">
                <div class="p-6">
                    <table class="min-w-full divide-y divide-gray-200">
                        <thead class="bg-gray-50">
                            <tr>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">ID</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
                                <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Created</th>
                                <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
                            </tr>
                        </thead>
                        <tbody class="bg-white divide-y divide-gray-200">
                            <!-- HTMX will load data here -->
                            <tr hx-get="/api/%s" hx-trigger="load" hx-swap="outerHTML">
                                <td colspan="4" class="px-6 py-4 text-center text-gray-500">Loading...</td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, name, name, pluralName, lowerName, pluralName)
}

func (g *Generator) generateShowView(name, lowerName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>View %s - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 View %s</h1>
                    </div>
                    <div class="flex items-center space-x-4">
                        <a href="/%s" class="text-gray-600 hover:text-gray-900">Back to List</a>
                    </div>
                </div>
            </div>
        </nav>
        
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div id="%s-detail" class="bg-white rounded-lg shadow p-6">
                <div class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700">ID</label>
                        <p class="mt-1 text-sm text-gray-900">1</p>
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Name</label>
                        <p class="mt-1 text-sm text-gray-900">Sample %s</p>
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Created At</label>
                        <p class="mt-1 text-sm text-gray-900">2024-01-01</p>
                    </div>
                    <div class="flex space-x-4">
                        <a href="/%s/edit" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">Edit</a>
                        <button hx-delete="/api/%s/{id}" 
                                hx-confirm="Are you sure you want to delete this?"
                                hx-target="#%s-detail"
                                class="bg-red-500 text-white px-4 py-2 rounded hover:bg-red-600">
                            Delete
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</body>
</html>`, name, name, lowerName, lowerName, name, lowerName, lowerName, lowerName)
}

func (g *Generator) generateCreateView(name, lowerName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Create %s - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Create %s</h1>
                    </div>
                    <div class="flex items-center">
                        <a href="/%s" class="text-gray-600 hover:text-gray-900">Back to List</a>
                    </div>
                </div>
            </div>
        </nav>
        
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div class="bg-white rounded-lg shadow p-6">
                <form hx-post="/api/%s" hx-target="#result" class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Name</label>
                        <input type="text" name="name" required 
                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Description</label>
                        <textarea name="description" rows="4"
                                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"></textarea>
                    </div>
                    <div class="flex space-x-4">
                        <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
                            Create
                        </button>
                        <a href="/%s" class="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600">
                            Cancel
                        </a>
                    </div>
                </form>
                <div id="result" class="mt-4"></div>
            </div>
        </div>
    </div>
</body>
</html>`, name, name, lowerName, lowerName, lowerName)
}

func (g *Generator) generateEditView(name, lowerName string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Edit %s - Dolphin Framework</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-100">
    <div class="min-h-screen">
        <nav class="bg-white shadow">
            <div class="max-w-7xl mx-auto px-4">
                <div class="flex justify-between h-16">
                    <div class="flex items-center">
                        <h1 class="text-xl font-semibold">🐬 Edit %s</h1>
                    </div>
                    <div class="flex items-center">
                        <a href="/%s" class="text-gray-600 hover:text-gray-900">Back to List</a>
                    </div>
                </div>
            </div>
        </nav>
        
        <div class="max-w-7xl mx-auto py-6 px-4">
            <div class="bg-white rounded-lg shadow p-6">
                <form hx-put="/api/%s/{id}" hx-target="#result" class="space-y-4">
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Name</label>
                        <input type="text" name="name" required 
                               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
                    </div>
                    <div>
                        <label class="block text-sm font-medium text-gray-700">Description</label>
                        <textarea name="description" rows="4"
                                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"></textarea>
                    </div>
                    <div class="flex space-x-4">
                        <button type="submit" class="bg-blue-500 text-white px-4 py-2 rounded hover:bg-blue-600">
                            Update
                        </button>
                        <a href="/%s" class="bg-gray-500 text-white px-4 py-2 rounded hover:bg-gray-600">
                            Cancel
                        </a>
                    </div>
                </form>
                <div id="result" class="mt-4"></div>
            </div>
        </div>
    </div>
</body>
</html>`, name, name, lowerName, lowerName, lowerName)
}

func (g *Generator) generateFormPartial(name, lowerName string) string {
	return fmt.Sprintf(`<div class="space-y-4">
    <div>
        <label class="block text-sm font-medium text-gray-700">Name</label>
        <input type="text" name="name" required 
               class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500">
    </div>
    <div>
        <label class="block text-sm font-medium text-gray-700">Description</label>
        <textarea name="description" rows="4"
                  class="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"></textarea>
    </div>
</div>`)
}

// generateRepositoryContent generates repository template
func (g *Generator) generateRepositoryContent(name string) string {
	lowerName := strings.ToLower(name)
	return fmt.Sprintf(`package repositories

import (
	"github.com/mrhoseah/dolphin/app/models"
	"gorm.io/gorm"
)

// %sRepository handles data access for %s
type %sRepository struct {
	db *gorm.DB
}

// New%sRepository creates a new %s repository
func New%sRepository(db *gorm.DB) *%sRepository {
	return &%sRepository{db: db}
}

// FindAll retrieves all %s
func (r *%sRepository) FindAll() ([]models.%s, error) {
	var items []models.%s
	err := r.db.Find(&items).Error
	return items, err
}

// FindByID retrieves a %s by ID
func (r *%sRepository) FindByID(id uint) (*models.%s, error) {
	var item models.%s
	err := r.db.First(&item, id).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

// Create creates a new %s
func (r *%sRepository) Create(item *models.%s) error {
	return r.db.Create(item).Error
}

// Update updates a %s
func (r *%sRepository) Update(item *models.%s) error {
	return r.db.Save(item).Error
}

// Delete deletes a %s by ID
func (r *%sRepository) Delete(id uint) error {
	return r.db.Delete(&models.%s{}, id).Error
}

// FindWhere finds %s with custom conditions
func (r *%sRepository) FindWhere(conditions map[string]interface{}) ([]models.%s, error) {
	var items []models.%s
	query := r.db
	for key, value := range conditions {
		query = query.Where(key, value)
	}
	err := query.Find(&items).Error
	return items, err
}

// Count returns the total count of %s
func (r *%sRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&models.%s{}).Count(&count).Error
	return count, err
}

// Paginate retrieves paginated %s
func (r *%sRepository) Paginate(page, pageSize int) ([]models.%s, int64, error) {
	var items []models.%s
	var total int64

	offset := (page - 1) * pageSize

	if err := r.db.Model(&models.%s{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.Offset(offset).Limit(pageSize).Find(&items).Error
	return items, total, err
}
`, name, lowerName, name, name, name, name, name, name,
		lowerName, name, name, name,
		lowerName, name, name, name,
		lowerName, name, name,
		lowerName, name, name,
		lowerName, name, name,
		lowerName, name, name, name,
		lowerName, name, name,
		lowerName, name, name, name, name)
}

// generateAPIControllerContent generates API controller template
func (g *Generator) generateAPIControllerContent(name string) string {
	lowerName := strings.ToLower(name)
	pluralName := lowerName + "s"
	return fmt.Sprintf(`package api

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/mrhoseah/dolphin/app/models"
	"github.com/mrhoseah/dolphin/app/repositories"
	"gorm.io/gorm"
)

// %sController handles API requests for %s
type %sController struct {
	repo *repositories.%sRepository
}

// New%sController creates a new %s API controller
func New%sController(db *gorm.DB) *%sController {
	return &%sController{
		repo: repositories.New%sRepository(db),
	}
}

// Index handles GET /api/%s
// @Summary List all %s
// @Description Get a list of all %s
// @Tags %s
// @Accept json
// @Produce json
// @Success 200 {array} models.%s
// @Router /api/%s [get]
func (c *%sController) Index(w http.ResponseWriter, r *http.Request) {
	items, err := c.repo.FindAll()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to retrieve %s"})
		return
	}
	render.JSON(w, r, items)
}

// Show handles GET /api/%s/{id}
// @Summary Get %s by ID
// @Description Get a single %s by ID
// @Tags %s
// @Accept json
// @Produce json
// @Param id path int true "%s ID"
// @Success 200 {object} models.%s
// @Failure 404 {object} map[string]string
// @Router /api/%s/{id} [get]
func (c *%sController) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

	item, err := c.repo.FindByID(uint(id))
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "%s not found"})
		return
	}

	render.JSON(w, r, item)
}

// Store handles POST /api/%s
// @Summary Create %s
// @Description Create a new %s
// @Tags %s
// @Accept json
// @Produce json
// @Param %s body models.%s true "%s data"
// @Success 201 {object} models.%s
// @Failure 400 {object} map[string]string
// @Router /api/%s [post]
func (c *%sController) Store(w http.ResponseWriter, r *http.Request) {
	var item models.%s
	if err := render.DecodeJSON(r.Body, &item); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := c.repo.Create(&item); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to create %s"})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, item)
}

// Update handles PUT /api/%s/{id}
// @Summary Update %s
// @Description Update an existing %s
// @Tags %s
// @Accept json
// @Produce json
// @Param id path int true "%s ID"
// @Param %s body models.%s true "%s data"
// @Success 200 {object} models.%s
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/%s/{id} [put]
func (c *%sController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

	item, err := c.repo.FindByID(uint(id))
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "%s not found"})
		return
	}

	if err := render.DecodeJSON(r.Body, item); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

	if err := c.repo.Update(item); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to update %s"})
		return
	}

	render.JSON(w, r, item)
}

// Destroy handles DELETE /api/%s/{id}
// @Summary Delete %s
// @Description Delete a %s by ID
// @Tags %s
// @Accept json
// @Produce json
// @Param id path int true "%s ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/%s/{id} [delete]
func (c *%sController) Destroy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

	if err := c.repo.Delete(uint(id)); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to delete %s"})
		return
	}

	render.JSON(w, r, map[string]string{"message": "%s deleted successfully"})
}
`, name, lowerName, name, name, name, name, name, name, name, name,
		pluralName, lowerName, lowerName, lowerName, name, pluralName, name, lowerName,
		pluralName, lowerName, lowerName, lowerName, name, name, pluralName, name,
		name, pluralName, lowerName, lowerName, lowerName, lowerName, name, name, name, pluralName, name, name,
		lowerName, pluralName, lowerName, lowerName, lowerName, name, lowerName, name, name, pluralName, name,
		name, name, pluralName, lowerName, lowerName, lowerName, name, pluralName, name,
		lowerName, name)
}

// generateProviderContent generates service provider template
func (g *Generator) generateProviderContent(name, providerType string, priority int) string {
	lowerName := strings.ToLower(name)
	return `package providers

import (
	"github.com/mrhoseah/dolphin/internal/providers"
)

// ` + name + `Provider implements ` + providerType + ` functionality
type ` + name + `Provider struct {
	config ` + name + `Config
}

// ` + name + `Config holds configuration for ` + providerType + ` provider
type ` + name + `Config struct {
	// Add your configuration fields here
	Enabled bool
}

// New` + name + `Provider creates a new ` + name + ` provider
func New` + name + `Provider() providers.ServiceProvider {
	return &` + name + `Provider{
		config: ` + name + `Config{
			Enabled: true,
		},
	}
}

func (p *` + name + `Provider) Name() string {
	return "` + lowerName + `"
}

func (p *` + name + `Provider) Priority() int {
	return ` + fmt.Sprintf("%d", priority) + `
}

func (p *` + name + `Provider) Register() error {
	// Register services in the container
	// Example: container.Bind("` + lowerName + `", p)
	return nil
}

func (p *` + name + `Provider) Boot() error {
	// Initialize services after registration
	return nil
}

// Add your provider-specific methods here
func (p *` + name + `Provider) ExampleMethod() error {
	// Implement your provider logic
	return nil
}`
}

// generatePostmanCollectionContent creates Postman collection JSON
func (g *Generator) generatePostmanCollectionContent() string {
	return `{
	"info": {
		"_postman_id": "dolphin-framework-api",
		"name": "Dolphin Framework API",
		"description": "Complete API collection for Dolphin Framework testing",
		"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		"_exporter_id": "dolphin-framework"
	},
	"item": [
		{
			"name": "Authentication",
			"item": [
				{
					"name": "Login",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"email\": \"admin@example.com\",\n  \"password\": \"password\"\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/auth/login",
							"host": ["{{base_url}}"],
							"path": ["api", "auth", "login"]
						},
						"description": "Authenticate user and get access token"
					},
					"response": []
				},
				{
					"name": "Register",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"name\": \"John Doe\",\n  \"email\": \"john@example.com\",\n  \"password\": \"password123\",\n  \"password_confirmation\": \"password123\"\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/auth/register",
							"host": ["{{base_url}}"],
							"path": ["api", "auth", "register"]
						},
						"description": "Register a new user"
					},
					"response": []
				},
				{
					"name": "Get Current User",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/auth/me",
							"host": ["{{base_url}}"],
							"path": ["api", "auth", "me"]
						},
						"description": "Get current authenticated user information"
					},
					"response": []
				}
			],
			"description": "Authentication endpoints for user login, registration, and token management"
		},
		{
			"name": "Users",
			"item": [
				{
					"name": "Get All Users",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/users?page=1&limit=10",
							"host": ["{{base_url}}"],
							"path": ["api", "users"],
							"query": [
								{
									"key": "page",
									"value": "1"
								},
								{
									"key": "limit",
									"value": "10"
								}
							]
						},
						"description": "Get paginated list of users"
					},
					"response": []
				},
				{
					"name": "Get User by ID",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/users/1",
							"host": ["{{base_url}}"],
							"path": ["api", "users", "1"]
						},
						"description": "Get specific user by ID"
					},
					"response": []
				},
				{
					"name": "Create User",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"name\": \"Jane Doe\",\n  \"email\": \"jane@example.com\",\n  \"password\": \"password123\"\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/users",
							"host": ["{{base_url}}"],
							"path": ["api", "users"]
						},
						"description": "Create a new user"
					},
					"response": []
				},
				{
					"name": "Update User",
					"request": {
						"method": "PUT",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"name\": \"Jane Smith\",\n  \"email\": \"jane.smith@example.com\"\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/users/1",
							"host": ["{{base_url}}"],
							"path": ["api", "users", "1"]
						},
						"description": "Update user information"
					},
					"response": []
				},
				{
					"name": "Delete User",
					"request": {
						"method": "DELETE",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/users/1",
							"host": ["{{base_url}}"],
							"path": ["api", "users", "1"]
						},
						"description": "Delete a user"
					},
					"response": []
				}
			],
			"description": "User management endpoints for CRUD operations"
		},
		{
			"name": "Posts",
			"item": [
				{
					"name": "Get All Posts",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/posts?page=1&limit=10",
							"host": ["{{base_url}}"],
							"path": ["api", "posts"],
							"query": [
								{
									"key": "page",
									"value": "1"
								},
								{
									"key": "limit",
									"value": "10"
								}
							]
						},
						"description": "Get paginated list of posts"
					},
					"response": []
				},
				{
					"name": "Get Post by ID",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/posts/1",
							"host": ["{{base_url}}"],
							"path": ["api", "posts", "1"]
						},
						"description": "Get specific post by ID"
					},
					"response": []
				},
				{
					"name": "Create Post",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"title\": \"My First Post\",\n  \"content\": \"This is the content of my first post.\",\n  \"published\": true\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/posts",
							"host": ["{{base_url}}"],
							"path": ["api", "posts"]
						},
						"description": "Create a new post"
					},
					"response": []
				},
				{
					"name": "Update Post",
					"request": {
						"method": "PUT",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"title\": \"Updated Post Title\",\n  \"content\": \"Updated content for the post.\",\n  \"published\": false\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/posts/1",
							"host": ["{{base_url}}"],
							"path": ["api", "posts", "1"]
						},
						"description": "Update post information"
					},
					"response": []
				},
				{
					"name": "Delete Post",
					"request": {
						"method": "DELETE",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/posts/1",
							"host": ["{{base_url}}"],
							"path": ["api", "posts", "1"]
						},
						"description": "Delete a post"
					},
					"response": []
				}
			],
			"description": "Post management endpoints for CRUD operations"
		},
		{
			"name": "Storage",
			"item": [
				{
					"name": "Upload File",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"body": {
							"mode": "formdata",
							"formdata": [
								{
									"key": "file",
									"type": "file",
									"src": []
								},
								{
									"key": "path",
									"value": "uploads/",
									"type": "text"
								}
							]
						},
						"url": {
							"raw": "{{base_url}}/api/storage/upload",
							"host": ["{{base_url}}"],
							"path": ["api", "storage", "upload"]
						},
						"description": "Upload a file to storage"
					},
					"response": []
				},
				{
					"name": "Download File",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/storage/download/uploads/file.jpg",
							"host": ["{{base_url}}"],
							"path": ["api", "storage", "download", "uploads", "file.jpg"]
						},
						"description": "Download a file from storage"
					},
					"response": []
				},
				{
					"name": "List Files",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/storage/list/uploads/",
							"host": ["{{base_url}}"],
							"path": ["api", "storage", "list", "uploads", ""]
						},
						"description": "List files in storage directory"
					},
					"response": []
				},
				{
					"name": "Delete File",
					"request": {
						"method": "DELETE",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/storage/delete/uploads/file.jpg",
							"host": ["{{base_url}}"],
							"path": ["api", "storage", "delete", "uploads", "file.jpg"]
						},
						"description": "Delete a file from storage"
					},
					"response": []
				}
			],
			"description": "File storage management endpoints"
		},
		{
			"name": "Cache",
			"item": [
				{
					"name": "Get Cache Value",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/cache/user:123",
							"host": ["{{base_url}}"],
							"path": ["api", "cache", "user:123"]
						},
						"description": "Get a value from cache"
					},
					"response": []
				},
				{
					"name": "Set Cache Value",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"key\": \"user:123\",\n  \"value\": \"user data\",\n  \"ttl\": 3600\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/cache",
							"host": ["{{base_url}}"],
							"path": ["api", "cache"]
						},
						"description": "Set a value in cache with TTL"
					},
					"response": []
				},
				{
					"name": "Delete Cache Value",
					"request": {
						"method": "DELETE",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/cache/user:123",
							"host": ["{{base_url}}"],
							"path": ["api", "cache", "user:123"]
						},
						"description": "Delete a value from cache"
					},
					"response": []
				},
				{
					"name": "Clear Cache",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/cache/clear",
							"host": ["{{base_url}}"],
							"path": ["api", "cache", "clear"]
						},
						"description": "Clear all cache"
					},
					"response": []
				}
			],
			"description": "Cache management endpoints"
		},
		{
			"name": "Events",
			"item": [
				{
					"name": "Dispatch Event",
					"request": {
						"method": "POST",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							},
							{
								"key": "Content-Type",
								"value": "application/json"
							}
						],
						"body": {
							"mode": "raw",
							"raw": "{\n  \"event\": \"user.created\",\n  \"payload\": {\n    \"user_id\": 123,\n    \"email\": \"user@example.com\",\n    \"username\": \"john_doe\"\n  }\n}"
						},
						"url": {
							"raw": "{{base_url}}/api/events/dispatch",
							"host": ["{{base_url}}"],
							"path": ["api", "events", "dispatch"]
						},
						"description": "Dispatch an event"
					},
					"response": []
				},
				{
					"name": "Get Event History",
					"request": {
						"method": "GET",
						"header": [
							{
								"key": "Authorization",
								"value": "Bearer {{access_token}}"
							}
						],
						"url": {
							"raw": "{{base_url}}/api/events/history?event=user.created&limit=10",
							"host": ["{{base_url}}"],
							"path": ["api", "events", "history"],
							"query": [
								{
									"key": "event",
									"value": "user.created"
								},
								{
									"key": "limit",
									"value": "10"
								}
							]
						},
						"description": "Get event history"
					},
					"response": []
				}
			],
			"description": "Event management endpoints"
		}
	],
	"event": [
		{
			"listen": "prerequest",
			"script": {
				"type": "text/javascript",
				"exec": [
					"// Set base URL if not already set",
					"if (!pm.environment.get('base_url')) {",
					"    pm.environment.set('base_url', 'http://localhost:8080');",
					"}"
				]
			}
		},
		{
			"listen": "test",
			"script": {
				"type": "text/javascript",
				"exec": [
					"// Auto-extract tokens from login response",
					"if (pm.response.json() && pm.response.json().data) {",
					"    const data = pm.response.json().data;",
					"    if (data.access_token) {",
					"        pm.environment.set('access_token', data.access_token);",
					"    }",
					"    if (data.refresh_token) {",
					"        pm.environment.set('refresh_token', data.refresh_token);",
					"    }",
					"}"
				]
			}
		}
	],
	"variable": [
		{
			"key": "base_url",
			"value": "http://localhost:8080",
			"type": "string"
		},
		{
			"key": "access_token",
			"value": "",
			"type": "string"
		},
		{
			"key": "refresh_token",
			"value": "",
			"type": "string"
		}
	]
}`
}
