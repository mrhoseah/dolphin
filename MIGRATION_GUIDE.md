# 🔄 Migration Guide: Updating to New Project Structure

## 📋 **Overview**

This guide helps you migrate from the old Dolphin Framework project structure to the new simplified structure. The main changes are:

1. **Single `/views` directory** instead of `/resources/views` and `/ui/views`
2. **`.fin.html` file extensions** instead of `.fin`
3. **Local Tailwind CSS** instead of CDN dependency
4. **Enhanced authentication system**
5. **Built-in GORM migration system** instead of Raptor dependency

## 🚨 **Breaking Changes**

### **1. Directory Structure Changes**

#### **Before (Old Structure)**
```
my-app/
├── resources/
│   └── views/
│       ├── layouts/
│       ├── pages/
│       └── auth/
├── ui/
│   └── views/
│       ├── layouts/
│       ├── pages/
│       └── auth/
└── ...
```

#### **After (New Structure)**
```
my-app/
├── views/                  # Single views directory
│   ├── layouts/
│   ├── pages/
│   ├── auth/
│   ├── emails/
│   └── components/
└── ...
```

### **2. File Extension Changes**

#### **Before**
- `login.fin`
- `register.fin`
- `dashboard.fin`

#### **After**
- `login.fin.html`
- `register.fin.html`
- `dashboard.fin.html`

### **3. Migration System Changes**

#### **Before (Raptor)**
```go
package migrations

import (
    raptor "github.com/mrhoseah/raptor/core"
)

type CreateUsersTable struct{}

func (m *CreateUsersTable) Up(s raptor.Schema) error {
    return s.CreateTable("users", []string{"id", "name", "email"})
}

func (m *CreateUsersTable) Down(s raptor.Schema) error {
    return s.DropTable("users")
}
```

#### **After (GORM)**
```go
package migrations

import (
    "gorm.io/gorm"
)

type CreateUsersTable struct{}

func (m *CreateUsersTable) Name() string {
    return "create_users_table"
}

func (m *CreateUsersTable) Up(db *gorm.DB) error {
    return db.Exec(`
        CREATE TABLE users (
            id BIGINT PRIMARY KEY AUTO_INCREMENT,
            name VARCHAR(255) NOT NULL,
            email VARCHAR(255) UNIQUE NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
        )
    `).Error
}

func (m *CreateUsersTable) Down(db *gorm.DB) error {
    return db.Exec("DROP TABLE users").Error
}
```

## 🔧 **Migration Steps**

### **Step 1: Update Directory Structure**

1. **Create new views directory**:
   ```bash
   mkdir -p views/{layouts,pages,auth,emails,components}
   ```

2. **Move template files**:
   ```bash
   # Move from resources/views to views/
   mv resources/views/* views/
   
   # Move from ui/views to views/ (if different)
   mv ui/views/* views/
   ```

3. **Remove old directories**:
   ```bash
   rm -rf resources/views
   rm -rf ui/views
   ```

### **Step 2: Rename Template Files**

1. **Rename all `.fin` files to `.fin.html`**:
   ```bash
   find views/ -name "*.fin" -exec sh -c 'mv "$1" "${1%.fin}.fin.html"' _ {} \;
   ```

2. **Update template references**:
   ```bash
   # Update @extends calls
   sed -i 's/@extends('\''layouts\/app'\'')/@extends('\''layouts\/app.fin.html'\'')/g' views/**/*.fin.html
   ```

### **Step 3: Update Migration Files**

1. **Update migration imports**:
   ```bash
   # Replace Raptor imports with GORM
   find database/migrations/ -name "*.go" -exec sed -i 's/raptor "github.com\/mrhoseah\/raptor\/core"/"gorm.io\/gorm"/g' {} \;
   ```

2. **Update migration methods**:
   ```bash
   # Update method signatures
   find database/migrations/ -name "*.go" -exec sed -i 's/func (m \*.*) Up(s raptor.Schema) error/func (m *\1) Up(db *gorm.DB) error/g' {} \;
   find database/migrations/ -name "*.go" -exec sed -i 's/func (m \*.*) Down(s raptor.Schema) error/func (m *\1) Down(db *gorm.DB) error/g' {} \;
   ```

3. **Update migration logic**:
   - Replace `s.CreateTable()` with `db.Exec()` SQL statements
   - Replace `s.DropTable()` with `db.Exec("DROP TABLE ...")`
   - Add proper error handling with `.Error`

### **Step 4: Update Configuration Files**

1. **Update `config/template.yaml`**:
   ```yaml
   template:
     layouts_dir: "views/layouts"
     partials_dir: "views/partials"
     pages_dir: "views/pages"
     components_dir: "views/components"
     emails_dir: "views/emails"
     extension: ".fin.html"  # Changed from ".html"
     auto_reload: true
     cache_templates: true
     default_layout: "app"
     enable_helpers: true
     escape_html: true
   ```

2. **Update `tailwind.config.js`**:
   ```javascript
   module.exports = {
     content: [
       "./views/**/*.{fin,fin.html,html,go}",  // Updated path
       "./public/**/*.html",
       "./internal/**/*.go",
     ],
     // ... rest of config
   }
   ```

### **Step 5: Update Go Code**

1. **Update template rendering calls**:
   ```go
   // Before
   c.HTML(http.StatusOK, "auth/login", data)
   
   // After
   c.HTML(http.StatusOK, "auth/login.fin.html", data)
   ```

2. **Update configuration in Go code**:
   ```go
   config := &template.Config{
       LayoutsDir:     "views/layouts",     // Updated path
       PartialsDir:    "views/partials",   // Updated path
       PagesDir:       "views/pages",       // Updated path
       ComponentsDir:  "views/components", // Updated path
       EmailsDir:      "views/emails",     // Updated path
       Extension:      ".fin.html",        // Updated extension
       AutoReload:     true,
       CacheTemplates: true,
   }
   ```

### **Step 6: Update Package.json**

1. **Add Tailwind CSS build scripts**:
   ```json
   {
     "name": "my-app",
     "version": "1.0.0",
     "scripts": {
       "build": "tailwindcss -i ./assets/css/global.css -o ./public/css/app.css --watch",
       "build:prod": "tailwindcss -i ./assets/css/global.css -o ./public/css/app.css --minify",
       "dev": "tailwindcss -i ./assets/css/global.css -o ./public/css/app.css --watch"
     },
     "devDependencies": {
       "tailwindcss": "^3.4.0",
       "@tailwindcss/forms": "^0.5.7",
       "@tailwindcss/typography": "^0.5.10"
     }
   }
   ```

2. **Create `assets/css/global.css`**:
   ```css
   @tailwind base;
   @tailwind components;
   @tailwind utilities;
   
   /* Your custom styles here */
   ```

### **Step 7: Update Template References**

1. **Update layout references in templates**:
   ```fin
   @extends('layouts/app.fin.html')
   ```

2. **Update component references**:
   ```fin
   @component('components/alert.fin.html')
   ```

## 🧪 **Testing Your Migration**

### **1. Verify File Structure**
```bash
# Check that all files are in the right place
ls -la views/
ls -la views/auth/
ls -la views/layouts/
```

### **2. Test Template Rendering**
```bash
# Start your application
dolphin serve

# Visit your pages to ensure templates render correctly
curl http://localhost:8080/login
curl http://localhost:8080/register
```

### **3. Test CSS Build**
```bash
# Install dependencies
npm install

# Build CSS
npm run build

# Check that CSS file is generated
ls -la public/css/app.css
```

## 🚀 **New Features Available**

After migration, you can take advantage of new features:

### **1. Enhanced Authentication**
```bash
# Generate complete auth system
dolphin make:auth
```

### **2. Local Tailwind CSS**
```bash
# Build CSS locally (no CDN dependency)
npm run build:prod
```

### **3. Better IDE Support**
- `.fin.html` files get proper syntax highlighting
- Better autocomplete and IntelliSense
- HTML/CSS linters work correctly

## 🔍 **Troubleshooting**

### **Common Issues**

1. **Templates not found**:
   - Check that files are in `/views` directory
   - Verify file extensions are `.fin.html`
   - Update template references in Go code

2. **CSS not loading**:
   - Run `npm install` to install dependencies
   - Run `npm run build` to compile CSS
   - Check that `public/css/app.css` exists

3. **Import path errors**:
   - Update Go import paths to use correct module name
   - Run `go mod tidy` to update dependencies

### **Getting Help**

If you encounter issues during migration:

1. **Check the logs**: Look for error messages in the console
2. **Verify file paths**: Ensure all files are in the correct locations
3. **Test incrementally**: Migrate one component at a time
4. **Use the CLI**: Run `dolphin new test-app` to see the new structure

## 📚 **Additional Resources**

- **Project Structure Guide**: `PROJECT_STRUCTURE.md`
- **Fin Template Documentation**: `FIN_TEMPLATE_SUMMARY.md`
- **API Documentation**: `README.md`
- **Changelog**: `CHANGELOG.md`

---

**Migration Complete!** 🎉

Your Dolphin Framework project now uses the new simplified structure with better IDE support and enhanced features.

*Need help? Check the documentation or create an issue on GitHub.*
