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
