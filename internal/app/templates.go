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
    "dolphin/app/models"
    "gorm.io/gorm"
)

// %[1]sRepository handles data access for %[2]s
type %[1]sRepository struct {
    db *gorm.DB
}

// New%[1]sRepository creates a new %[1]s repository
func New%[1]sRepository(db *gorm.DB) *%[1]sRepository {
    return &%[1]sRepository{db: db}
}

func (r *%[1]sRepository) FindAll() ([]models.%[1]s, error) {
    var items []models.%[1]s
    err := r.db.Find(&items).Error
    return items, err
}

func (r *%[1]sRepository) FindByID(id uint) (*models.%[1]s, error) {
    var item models.%[1]s
    if err := r.db.First(&item, id).Error; err != nil {
        return nil, err
    }
    return &item, nil
}

func (r *%[1]sRepository) Create(item *models.%[1]s) error { return r.db.Create(item).Error }
func (r *%[1]sRepository) Update(item *models.%[1]s) error { return r.db.Save(item).Error }
func (r *%[1]sRepository) Delete(id uint) error { return r.db.Delete(&models.%[1]s{}, id).Error }

func (r *%[1]sRepository) Count() (int64, error) {
    var count int64
    err := r.db.Model(&models.%[1]s{}).Count(&count).Error
    return count, err
}

func (r *%[1]sRepository) Paginate(page, pageSize int) ([]models.%[1]s, int64, error) {
    var items []models.%[1]s
    var total int64
    offset := (page - 1) * pageSize
    if err := r.db.Model(&models.%[1]s{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }
    err := r.db.Offset(offset).Limit(pageSize).Find(&items).Error
    return items, total, err
}
`, name, lowerName)
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
	"dolphin/app/models"
	"dolphin/app/repositories"
	"gorm.io/gorm"
)

// %[1]sController handles API requests for %[2]s
type %[1]sController struct { repo *repositories.%[1]sRepository }

// New%[1]sController creates a new %[1]s API controller
func New%[1]sController(db *gorm.DB) *%[1]sController {
    return &%[1]sController{ repo: repositories.New%[1]sRepository(db) }
}

// Index handles GET /api/%[3]s
// @Summary List all %[3]s
// @Description Get a list of all %[3]s
// @Tags %[1]s
// @Accept json
// @Produce json
// @Success 200 {array} models.%[1]s
// @Router /api/%[3]s [get]
func (c *%[1]sController) Index(w http.ResponseWriter, r *http.Request) {
	items, err := c.repo.FindAll()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, map[string]string{"error": "Failed to retrieve %[2]s"})
		return
	}
	render.JSON(w, r, items)
}

// Show handles GET /api/%[3]s/{id}
// @Summary Get %[2]s by ID
// @Description Get a single %[2]s by ID
// @Tags %[1]s
// @Accept json
// @Produce json
// @Param id path int true "%[2]s ID"
// @Success 200 {object} models.%[1]s
// @Failure 404 {object} map[string]string
// @Router /api/%[3]s/{id} [get]
func (c *%[1]sController) Show(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

    item, err := c.repo.FindByID(uint(id))
	if err != nil {
		render.Status(r, http.StatusNotFound)
        render.JSON(w, r, map[string]string{"error": "%[2]s not found"})
		return
	}

	render.JSON(w, r, item)
}

// Store handles POST /api/%[3]s
// @Summary Create %[2]s
// @Description Create a new %[2]s
// @Tags %[1]s
// @Accept json
// @Produce json
// @Param %[2]s body models.%[1]s true "%[2]s data"
// @Success 201 {object} models.%[1]s
// @Failure 400 {object} map[string]string
// @Router /api/%[3]s [post]
func (c *%[1]sController) Store(w http.ResponseWriter, r *http.Request) {
    var item models.%[1]s
	if err := render.DecodeJSON(r.Body, &item); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

    if err := c.repo.Create(&item); err != nil {
		render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, map[string]string{"error": "Failed to create %[2]s"})
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, item)
}

// Update handles PUT /api/%[3]s/{id}
// @Summary Update %[2]s
// @Description Update an existing %[2]s
// @Tags %[1]s
// @Accept json
// @Produce json
// @Param id path int true "%[2]s ID"
// @Param %[2]s body models.%[1]s true "%[2]s data"
// @Success 200 {object} models.%[1]s
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/%[3]s/{id} [put]
func (c *%[1]sController) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

    item, err := c.repo.FindByID(uint(id))
	if err != nil {
		render.Status(r, http.StatusNotFound)
        render.JSON(w, r, map[string]string{"error": "%[2]s not found"})
		return
	}

	if err := render.DecodeJSON(r.Body, item); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}

    if err := c.repo.Update(item); err != nil {
		render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, map[string]string{"error": "Failed to update %[2]s"})
		return
	}

	render.JSON(w, r, item)
}

// Destroy handles DELETE /api/%[3]s/{id}
// @Summary Delete %[2]s
// @Description Delete a %[2]s by ID
// @Tags %[1]s
// @Accept json
// @Produce json
// @Param id path int true "%[2]s ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/%[3]s/{id} [delete]
func (c *%[1]sController) Destroy(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid ID"})
		return
	}

    if err := c.repo.Delete(uint(id)); err != nil {
		render.Status(r, http.StatusInternalServerError)
        render.JSON(w, r, map[string]string{"error": "Failed to delete %[2]s"})
		return
	}

    render.JSON(w, r, map[string]string{"message": "%[2]s deleted successfully"})
}
`, name, lowerName, pluralName)
}

// generateProviderContent generates service provider template
func (g *Generator) generateProviderContent(name, providerType string, priority int) string {
	lowerName := strings.ToLower(name)
	return `package providers

import (
	"dolphin/internal/providers"
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

// generateAuthViewContent generates authentication view templates
func (g *Generator) generateAuthViewContent(viewType string) string {
	switch viewType {
	case "login":
		return g.generateLoginView()
	case "register":
		return g.generateRegisterView()
	case "forgot-password":
		return g.generateForgotPasswordView()
	case "reset-password":
		return g.generateResetPasswordView()
	default:
		return ""
	}
}

// generateLoginView generates the login page template
func (g *Generator) generateLoginView() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "content"}}
<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Sign in to your account
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/auth/login" hx-target="#error-message" hx-swap="outerHTML">
            <div class="rounded-md shadow-sm -space-y-px">
                <div>
                    <label for="email" class="sr-only">Email address</label>
                    <input id="email" name="email" type="email" autocomplete="email" required
                           class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Email address">
                </div>
                <div>
                    <label for="password" class="sr-only">Password</label>
                    <input id="password" name="password" type="password" autocomplete="current-password" required
                           class="appearance-none rounded-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Password">
                </div>
            </div>

            <div class="flex items-center justify-between">
                <div class="flex items-center">
                    <input id="remember-me" name="remember-me" type="checkbox"
                           class="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded">
                    <label for="remember-me" class="ml-2 block text-sm text-gray-900">
                        Remember me
                    </label>
                </div>

                <div class="text-sm">
                    <a href="/auth/forgot-password" class="font-medium text-blue-600 hover:text-blue-500">
                        Forgot your password?
                    </a>
                </div>
            </div>

            <div id="error-message">
                {{if .Error}}
                    <div class="rounded-md bg-red-50 p-4 mb-4">
                        <div class="text-sm text-red-800">{{.Error}}</div>
                    </div>
                {{end}}
            </div>

            <div>
                <button type="submit"
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Sign in
                </button>
            </div>

            <div class="text-center">
                <span class="text-sm text-gray-600">
                    Don't have an account?
                    <a href="/auth/register" class="font-medium text-blue-600 hover:text-blue-500">
                        Sign up
                    </a>
                </span>
            </div>
        </form>
    </div>
</div>
{{end}}`
}

// generateRegisterView generates the register page template
func (g *Generator) generateRegisterView() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "content"}}
<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Create your account
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/auth/register" hx-target="#error-message" hx-swap="outerHTML">
            <div class="rounded-md shadow-sm space-y-4">
                <div>
                    <label for="name" class="block text-sm font-medium text-gray-700">Name</label>
                    <input id="name" name="name" type="text" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Full name">
                </div>
                <div>
                    <label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
                    <input id="email" name="email" type="email" autocomplete="email" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Email address">
                </div>
                <div>
                    <label for="password" class="block text-sm font-medium text-gray-700">Password</label>
                    <input id="password" name="password" type="password" autocomplete="new-password" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Password">
                </div>
                <div>
                    <label for="password-confirm" class="block text-sm font-medium text-gray-700">Confirm Password</label>
                    <input id="password-confirm" name="password-confirm" type="password" autocomplete="new-password" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Confirm password">
                </div>
            </div>

            <div id="error-message">
                {{if .Error}}
                    <div class="rounded-md bg-red-50 p-4 mb-4">
                        <div class="text-sm text-red-800">{{.Error}}</div>
                    </div>
                {{end}}
            </div>

            <div>
                <button type="submit"
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Sign up
                </button>
            </div>

            <div class="text-center">
                <span class="text-sm text-gray-600">
                    Already have an account?
                    <a href="/auth/login" class="font-medium text-blue-600 hover:text-blue-500">
                        Sign in
                    </a>
                </span>
            </div>
        </form>
    </div>
</div>
{{end}}`
}

// generateForgotPasswordView generates the forgot password page template
func (g *Generator) generateForgotPasswordView() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "content"}}
<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Forgot your password?
            </h2>
            <p class="mt-2 text-center text-sm text-gray-600">
                Enter your email address and we'll send you a link to reset your password.
            </p>
        </div>
        <form class="mt-8 space-y-6" hx-post="/auth/forgot-password" hx-target="#error-message" hx-swap="outerHTML">
            <div>
                <label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
                <input id="email" name="email" type="email" autocomplete="email" required
                       class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                       placeholder="Email address">
            </div>

            <div id="error-message"></div>

            <div>
                <button type="submit"
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Send reset link
                </button>
            </div>

            <div class="text-center">
                <a href="/auth/login" class="font-medium text-blue-600 hover:text-blue-500">
                    Back to login
                </a>
            </div>
        </form>
    </div>
</div>
{{end}}`
}

// generateResetPasswordView generates the reset password page template
func (g *Generator) generateResetPasswordView() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "content"}}
<div class="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
    <div class="max-w-md w-full space-y-8">
        <div>
            <h2 class="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Reset your password
            </h2>
        </div>
        <form class="mt-8 space-y-6" hx-post="/auth/reset-password" hx-target="#error-message" hx-swap="outerHTML">
            <input type="hidden" name="token" value="{{.Token}}">
            
            <div class="rounded-md shadow-sm space-y-4">
                <div>
                    <label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
                    <input id="email" name="email" type="email" autocomplete="email" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Email address">
                </div>
                <div>
                    <label for="password" class="block text-sm font-medium text-gray-700">New Password</label>
                    <input id="password" name="password" type="password" autocomplete="new-password" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="New password">
                </div>
                <div>
                    <label for="password-confirm" class="block text-sm font-medium text-gray-700">Confirm Password</label>
                    <input id="password-confirm" name="password-confirm" type="password" autocomplete="new-password" required
                           class="mt-1 appearance-none relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
                           placeholder="Confirm password">
                </div>
            </div>

            <div id="error-message"></div>

            <div>
                <button type="submit"
                        class="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                    Reset password
                </button>
            </div>

            <div class="text-center">
                <a href="/auth/login" class="font-medium text-blue-600 hover:text-blue-500">
                    Back to login
                </a>
            </div>
        </form>
    </div>
</div>
{{end}}`
}

// generateBaseLayoutContent generates the base layout template
func (g *Generator) generateBaseLayoutContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Dolphin Framework{{end}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-50">
    <nav class="bg-white shadow-sm">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex items-center">
                    <a href="/" class="text-xl font-bold text-blue-600">🐬 Dolphin</a>
                </div>
                <div class="flex items-center space-x-4">
                    {{if .User}}
                        <span class="text-gray-700">{{.User.Email}}</span>
                        <form hx-post="/auth/logout" hx-swap="none">
                            <button type="submit" class="text-gray-600 hover:text-gray-900">
                                Logout
                            </button>
                        </form>
                    {{else}}
                        <a href="/auth/login" class="text-gray-600 hover:text-gray-900">Login</a>
                        <a href="/auth/register" class="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700">
                            Register
                        </a>
                    {{end}}
                </div>
            </div>
        </div>
    </nav>

    <main>
        {{block "content" .}}{{end}}
    </main>

    <footer class="bg-white border-t mt-auto">
        <div class="max-w-7xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
            <p class="text-center text-sm text-gray-500">
                &copy; {{.Year}} Dolphin Framework. All rights reserved.
            </p>
        </div>
    </footer>
</body>
</html>`
}

// generateBaseLayoutWithAuth generates base layout with enhanced auth navigation
func (g *Generator) generateBaseLayoutWithAuth() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Dolphin Framework{{end}}</title>
    <script src="https://cdn.tailwindcss.com"></script>
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-50">
    <nav class="bg-white shadow-sm border-b border-gray-200">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="flex justify-between h-16">
                <div class="flex items-center">
                    <a href="/" class="text-xl font-bold text-blue-600 hover:text-blue-700">
                        🐬 Dolphin
                    </a>
                </div>
                <div class="flex items-center space-x-4">
                    {{if .User}}
                        <!-- Authenticated User Menu -->
                        <div class="flex items-center space-x-4">
                            <span class="text-sm text-gray-700">{{.User.Name}}</span>
                            <a href="/dashboard" class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">
                                Dashboard
                            </a>
                            <form hx-post="/auth/logout" hx-swap="none" class="inline">
                                <button type="submit" class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">
                                    Logout
                                </button>
                            </form>
                        </div>
                    {{else}}
                        <!-- Guest Navigation -->
                        <a href="/auth/login" class="text-gray-600 hover:text-gray-900 px-3 py-2 rounded-md text-sm font-medium">
                            Log in
                        </a>
                        <a href="/auth/register" class="bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                            Register
                        </a>
                    {{end}}
                </div>
            </div>
        </div>
    </nav>

    <main class="min-h-screen">
        {{block "content" .}}{{end}}
    </main>

    <footer class="bg-white border-t mt-auto">
        <div class="max-w-7xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
            <p class="text-center text-sm text-gray-500">
                &copy; {{.Year}} Dolphin Framework. All rights reserved.
            </p>
        </div>
    </footer>
</body>
</html>`
}

// generateAuthControllerContent generates AuthController with web methods
func (g *Generator) generateAuthControllerContent(moduleName string) string {
	return `package controllers

import (
	"net/http"
	"` + moduleName + `/app/models"
	"dolphin/pkg/auth"
	"dolphin/pkg/template"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthController handles authentication web requests
type AuthController struct {
	authManager *auth.AuthManager
	template    template.FinTemplateEngine
	logger      *zap.Logger
	db          *gorm.DB
}

// NewAuthController creates a new AuthController
func NewAuthController(authManager *auth.AuthManager, templateEngine template.FinTemplateEngine, logger *zap.Logger, db *gorm.DB) *AuthController {
	return &AuthController{
		authManager: authManager,
		template:    templateEngine,
		logger:      logger,
		db:          db,
	}
}

// ShowLogin displays the login page
func (c *AuthController) ShowLogin(w http.ResponseWriter, r *http.Request) {
	// Check if already authenticated
	if c.authManager.Check() {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"User": nil,
	}

	html, err := c.template.Render("auth/login", data)
	if err != nil {
		c.logger.Error("Failed to render login page", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleLogin processes login form submission
func (c *AuthController) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		data := map[string]interface{}{
			"User": nil,
			"Error": "Email and password are required",
		}
		html, _ := c.template.Render("auth/login", data)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(html))
		return
	}

	// Attempt authentication
	success, err := c.authManager.Attempt(map[string]string{
		"email":    email,
		"password": password,
	})

	if !success || err != nil {
		data := map[string]interface{}{
			"User": nil,
			"Error": "Invalid credentials",
		}
		html, _ := c.template.Render("auth/login", data)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(html))
		return
	}

	// Redirect to dashboard on success
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// ShowRegister displays the registration page
func (c *AuthController) ShowRegister(w http.ResponseWriter, r *http.Request) {
	// Check if already authenticated
	if c.authManager.Check() {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}

	data := map[string]interface{}{
		"User": nil,
	}

	html, err := c.template.Render("auth/register", data)
	if err != nil {
		c.logger.Error("Failed to render register page", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleRegister processes registration form submission
func (c *AuthController) HandleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password-confirm")

	if name == "" || email == "" || password == "" || passwordConfirm == "" {
		data := map[string]interface{}{
			"User": nil,
			"Error": "All fields are required",
		}
		html, _ := c.template.Render("auth/register", data)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(html))
		return
	}

	if password != passwordConfirm {
		data := map[string]interface{}{
			"User": nil,
			"Error": "Passwords do not match",
		}
		html, _ := c.template.Render("auth/register", data)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(html))
		return
	}

	// Check if user already exists
	var existing models.User
	if err := c.db.Where("email = ?", email).First(&existing).Error; err == nil {
		data := map[string]interface{}{
			"User":  nil,
			"Error": "Email already registered",
		}
		html, _ := c.template.Render("auth/register", data)
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(html))
		return
	}

	// Create new user
	user := models.User{
		Email: email,
		Name:  name,
	}
	if err := user.SetPassword(password); err != nil {
		c.logger.Error("Failed to hash password", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if err := c.db.Create(&user).Error; err != nil {
		c.logger.Error("Failed to create user", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Auto-login after registration
	if err := c.authManager.Login(&user); err != nil {
		c.logger.Error("Auto-login failed after registration", zap.Error(err))
		http.Redirect(w, r, "/auth/login", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// HandleLogout processes logout
func (c *AuthController) HandleLogout(w http.ResponseWriter, r *http.Request) {
	c.authManager.Logout()
	http.Redirect(w, r, "/", http.StatusFound)
}

// ShowForgotPassword displays the forgot password page
func (c *AuthController) ShowForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"User": nil,
	}

	html, err := c.template.Render("auth/forgot-password", data)
	if err != nil {
		c.logger.Error("Failed to render forgot password page", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(html))
}

// HandleForgotPassword processes forgot password form
func (c *AuthController) HandleForgotPassword(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement password reset email sending
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}
`
}

// generateAuthRoutesContent generates auth routes setup file
func (g *Generator) generateAuthRoutesContent(moduleName string) string {
	return `package bootstrap

import (
	"` + moduleName + `/app/http/controllers"
	"` + moduleName + `/app/models"
	"dolphin/pkg/auth"
	"dolphin/pkg/template"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"github.com/go-chi/chi/v5"
	dolphinMiddleware "dolphin/pkg/middleware"
)

// SetupAuthRoutes configures authentication routes
// Call this function from your main routes setup
func SetupAuthRoutes(r chi.Router, authManager *auth.AuthManager, templateEngine template.FinTemplateEngine, logger *zap.Logger, db *gorm.DB) {
	// IMPORTANT: Initialize auth guards before using auth middleware
	// Without guards, authManager.Check() will panic with nil pointer
	
	// Create a session store (in-memory for now, can be replaced with database-backed)
	sessionStore := auth.NewMemorySessionStore()
	
	// Create user provider (uses User model from database)
	userProvider := auth.NewDatabaseProvider(db, &models.User{})
	
	// Create and register session guard for web authentication
	sessionGuard := auth.NewSessionGuard("web", userProvider, sessionStore)
	authManager.RegisterGuard("web", sessionGuard)
	authManager.SetDefaultGuard("web")
	
	logger.Info("Auth guards initialized", zap.String("guard", "web"))

	// Create auth controller
	authController := controllers.NewAuthController(authManager, templateEngine, logger, db)

	// Create auth middleware
	authMiddleware := dolphinMiddleware.NewAuthMiddleware(authManager, logger)

	// Authentication routes
	r.Route("/auth", func(auth chi.Router) {
		// Public routes (guest only)
		auth.Group(func(guest chi.Router) {
			guest.Use(authMiddleware.Guest)
			guest.Get("/login", authController.ShowLogin)
			guest.Post("/login", authController.HandleLogin)
			guest.Get("/register", authController.ShowRegister)
			guest.Post("/register", authController.HandleRegister)
			guest.Get("/forgot-password", authController.ShowForgotPassword)
			guest.Post("/forgot-password", authController.HandleForgotPassword)
		})

		// Protected routes (authenticated only)
		auth.Group(func(protected chi.Router) {
			protected.Use(authMiddleware.Authenticate)
			protected.Post("/logout", authController.HandleLogout)
		})
	})
}
`
}

// generateHomePageContent generates the home page template
func (g *Generator) generateHomePageContent() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "content"}}
<div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
        <div class="text-center">
            <h1 class="text-4xl font-extrabold text-gray-900 sm:text-5xl md:text-6xl">
                Welcome to
                <span class="text-blue-600">Your App</span>
            </h1>
            <p class="mt-3 max-w-md mx-auto text-base text-gray-500 sm:text-lg md:mt-5 md:text-xl md:max-w-3xl">
                Built with Dolphin Framework
            </p>
            
            <div class="mt-10 flex justify-center space-x-4">
                {{if .User}}
                    <a href="/dashboard" class="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Go to Dashboard
                    </a>
                {{else}}
                    <a href="/auth/login" class="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Sign In
                    </a>
                    <a href="/auth/register" class="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500">
                        Register
                    </a>
                {{end}}
            </div>
        </div>

        <div class="mt-20 grid grid-cols-1 gap-8 sm:grid-cols-2 lg:grid-cols-3">
            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-center h-12 w-12 rounded-md bg-blue-500 text-white">
                    <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path>
                    </svg>
                </div>
                <h3 class="mt-4 text-lg font-medium text-gray-900">Fast & Efficient</h3>
                <p class="mt-2 text-base text-gray-500">
                    Built with Go for high performance and scalability.
                </p>
            </div>

            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-center h-12 w-12 rounded-md bg-green-500 text-white">
                    <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
                    </svg>
                </div>
                <h3 class="mt-4 text-lg font-medium text-gray-900">Secure</h3>
                <p class="mt-2 text-base text-gray-500">
                    Built-in authentication and security features.
                </p>
            </div>

            <div class="bg-white rounded-lg shadow-md p-6">
                <div class="flex items-center justify-center h-12 w-12 rounded-md bg-purple-500 text-white">
                    <svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4"></path>
                    </svg>
                </div>
                <h3 class="mt-4 text-lg font-medium text-gray-900">Developer Friendly</h3>
                <p class="mt-2 text-base text-gray-500">
                    Easy to use and extend with modern tooling.
                </p>
            </div>
        </div>
    </div>
</div>
{{end}}
`
}

// generateDashboardPageContent generates the dashboard page template
func (g *Generator) generateDashboardPageContent() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "title"}}Dashboard{{end}}

{{define "content"}}
<div class="min-h-screen bg-gray-50">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div class="mb-8">
            <h1 class="text-3xl font-bold text-gray-900">Dashboard</h1>
            <p class="mt-2 text-sm text-gray-600">Welcome to your dashboard</p>
        </div>

        <!-- Stats Grid -->
        <div class="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="p-5">
                    <div class="flex items-center">
                        <div class="flex-shrink-0">
                            <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"></path>
                            </svg>
                        </div>
                        <div class="ml-5 w-0 flex-1">
                            <dl>
                                <dt class="text-sm font-medium text-gray-500 truncate">Total Users</dt>
                                <dd class="text-lg font-medium text-gray-900">0</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="p-5">
                    <div class="flex items-center">
                        <div class="flex-shrink-0">
                            <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
                            </svg>
                        </div>
                        <div class="ml-5 w-0 flex-1">
                            <dl>
                                <dt class="text-sm font-medium text-gray-500 truncate">Statistics</dt>
                                <dd class="text-lg font-medium text-gray-900">0</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="p-5">
                    <div class="flex items-center">
                        <div class="flex-shrink-0">
                            <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"></path>
                            </svg>
                        </div>
                        <div class="ml-5 w-0 flex-1">
                            <dl>
                                <dt class="text-sm font-medium text-gray-500 truncate">Events</dt>
                                <dd class="text-lg font-medium text-gray-900">0</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>

            <div class="bg-white overflow-hidden shadow rounded-lg">
                <div class="p-5">
                    <div class="flex items-center">
                        <div class="flex-shrink-0">
                            <svg class="h-6 w-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
                            </svg>
                        </div>
                        <div class="ml-5 w-0 flex-1">
                            <dl>
                                <dt class="text-sm font-medium text-gray-500 truncate">Documents</dt>
                                <dd class="text-lg font-medium text-gray-900">0</dd>
                            </dl>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Quick Actions -->
        <div class="bg-white shadow rounded-lg mb-8">
            <div class="px-6 py-4 border-b border-gray-200">
                <h2 class="text-lg font-medium text-gray-900">Quick Actions</h2>
            </div>
            <div class="p-6">
                <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    <a href="#" class="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
                        <svg class="h-6 w-6 text-blue-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path>
                        </svg>
                        <span class="text-gray-700 font-medium">Create New</span>
                    </a>
                    <a href="#" class="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
                        <svg class="h-6 w-6 text-green-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"></path>
                        </svg>
                        <span class="text-gray-700 font-medium">Upload</span>
                    </a>
                    <a href="#" class="flex items-center p-4 border border-gray-200 rounded-lg hover:bg-gray-50">
                        <svg class="h-6 w-6 text-purple-600 mr-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"></path>
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"></path>
                        </svg>
                        <span class="text-gray-700 font-medium">Settings</span>
                    </a>
                </div>
            </div>
        </div>

        <!-- Recent Activity -->
        <div class="bg-white shadow rounded-lg">
            <div class="px-6 py-4 border-b border-gray-200">
                <h2 class="text-lg font-medium text-gray-900">Recent Activity</h2>
            </div>
            <div class="p-6">
                <p class="text-gray-500 text-center py-8">No recent activity to display</p>
            </div>
        </div>
    </div>
</div>
{{end}}
`
}

// generateErrorPageContent generates the error page template
func (g *Generator) generateErrorPageContent() string {
	return `{{extend "layouts/base.fin.html"}}

{{define "title"}}{{if .Title}}{{.Title}}{{else}}Error{{end}}{{end}}

{{define "content"}}
<div class="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center p-4">
	<div class="bg-white rounded-2xl shadow-2xl max-w-3xl w-full p-8 md:p-12">
		<div class="text-center mb-8">
			{{if eq .StatusCode 404}}
				<div class="text-6xl md:text-8xl mb-4">🔍</div>
				<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">Page Not Found</h1>
				<p class="text-lg text-gray-600">The page you're looking for doesn't exist or has been moved.</p>
			{{else if eq .StatusCode 500}}
				<div class="text-6xl md:text-8xl mb-4">⚠️</div>
				<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">Internal Server Error</h1>
				<p class="text-lg text-gray-600">An unexpected error occurred while processing your request.</p>
			{{else if eq .StatusCode 403}}
				<div class="text-6xl md:text-8xl mb-4">🔒</div>
				<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">Access Forbidden</h1>
				<p class="text-lg text-gray-600">You don't have permission to access this resource.</p>
			{{else if eq .StatusCode 401}}
				<div class="text-6xl md:text-8xl mb-4">🔐</div>
				<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">Unauthorized</h1>
				<p class="text-lg text-gray-600">Please log in to access this resource.</p>
			{{else}}
				<div class="text-6xl md:text-8xl mb-4">🐬</div>
				<h1 class="text-3xl md:text-4xl font-bold text-gray-900 mb-2">{{.Title}}</h1>
				<p class="text-lg text-gray-600">{{.Message}}</p>
			{{end}}
		</div>
		
		{{if or .Debug (eq .Environment "development")}}
		<div class="bg-gray-50 border border-gray-200 rounded-lg p-6 mb-6">
			<h3 class="text-lg font-semibold text-gray-900 mb-4">Debug Information</h3>
			<div class="space-y-2 text-sm">
				<p><strong class="text-gray-700">Path:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{{.Path}}</code></p>
				<p><strong class="text-gray-700">Method:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{{.Method}}</code></p>
				<p><strong class="text-gray-700">Status Code:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{{.StatusCode}}</code></p>
				<p><strong class="text-gray-700">Environment:</strong> <code class="bg-gray-200 px-2 py-1 rounded">{{.Environment}}</code></p>
			</div>
		</div>
		{{end}}
		
		<div class="border-t border-gray-200 pt-6">
			<div class="flex flex-col sm:flex-row gap-4 justify-center">
				<a href="/" class="inline-flex items-center justify-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6"></path>
					</svg>
					Go Home
				</a>
				<button onclick="window.history.back()" class="inline-flex items-center justify-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 transition-colors">
					<svg class="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 19l-7-7m0 0l7-7m-7 7h18"></path>
					</svg>
					Go Back
				</button>
			</div>
		</div>
	</div>
</div>
{{end}}
`
}

// generateUserModelContent generates User model for authentication
func (g *Generator) generateUserModelContent() string {
	return `package models

import (
    "time"
    "golang.org/x/crypto/bcrypt"
)

type User struct {
    ID        uint      ` + "`gorm:\"primaryKey\" json:\"id\"`" + `
    Email     string    ` + "`gorm:\"unique;not null\" json:\"email\"`" + `
    Password  string    ` + "`gorm:\"not null\" json:\"-\"`" + ` // Hashed password
    Name      string    ` + "`gorm:\"not null\" json:\"name\"`" + `
    CreatedAt time.Time ` + "`json:\"created_at\"`" + `
    UpdatedAt time.Time ` + "`json:\"updated_at\"`" + `
}

// SetPassword hashes the password before storing
func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

// CheckPassword verifies the password
func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
`
}

// generateUsersMigrationContent generates a GORM migration for the users table
func (g *Generator) generateUsersMigrationContent(moduleName string) string {
	return `package migrations

import (
    "log"

    "` + moduleName + `/app/models"
    "gorm.io/gorm"
)

// UpUsers creates the users table
func UpUsers(db *gorm.DB) error {
    if err := db.AutoMigrate(&models.User{}); err != nil {
        return err
    }
    return nil
}

// DownUsers drops the users table
func DownUsers(db *gorm.DB) error {
    if err := db.Migrator().DropTable(&models.User{}); err != nil {
        // Log and continue to avoid blocking rollbacks
        log.Printf("drop users table failed: %v", err)
    }
    return nil
}
`
}
