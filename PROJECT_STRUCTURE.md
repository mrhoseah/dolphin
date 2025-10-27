# 🐬 Dolphin Framework - Project Structure

## 📁 **Simplified Directory Structure**

Dolphin Framework uses a clean, intuitive project structure that's easy to understand and navigate. Unlike other frameworks that have complex nested directories, Dolphin keeps things simple with a single `/views` directory for all templates.

### **Generated Project Structure**

When you run `dolphin new my-app`, you get this clean structure:

```
my-app/
├── app/                    # Application logic
│   ├── http/controllers/   # HTTP controllers
│   ├── http/middleware/    # Middleware components
│   ├── http/requests/      # Request validation
│   ├── models/            # Data models
│   ├── services/          # Business logic services
│   └── mail/              # Email services
├── views/                  # 🎯 Single views directory
│   ├── layouts/           # Base layouts (.fin.html)
│   ├── components/        # Reusable components
│   ├── pages/             # Page templates
│   ├── auth/              # Authentication views
│   └── emails/            # Email templates
├── assets/                 # Source assets
│   ├── css/               # CSS source files
│   └── js/                # JavaScript source files
├── public/                 # Compiled assets
│   ├── css/               # Compiled CSS
│   ├── js/                # Compiled JavaScript
│   └── images/            # Static images
├── database/               # Database files
│   ├── migrations/        # Database migrations
│   └── seeders/           # Database seeders
├── config/                 # Configuration files
├── routes/                 # Route definitions
├── storage/                # Storage directories
│   ├── logs/              # Application logs
│   ├── cache/             # Cache files
│   └── sessions/          # Session storage
├── tests/                  # Test files
├── internal/               # Internal packages
├── main.go                 # Application entry point
├── go.mod                  # Go module file
├── package.json            # Node.js dependencies
├── tailwind.config.js      # Tailwind CSS configuration
└── README.md               # Project documentation
```

## 🎯 **Key Changes from Previous Versions**

### **Before (Complex Structure)**
```
my-app/
├── resources/              # ❌ Confusing Laravel-like structure
│   └── views/             # ❌ Nested directories
├── ui/                    # ❌ Redundant directory
│   └── views/             # ❌ Duplicate views
└── ...
```

### **After (Simplified Structure)**
```
my-app/
├── views/                 # ✅ Single, clear directory
│   ├── layouts/           # ✅ Base layouts
│   ├── auth/              # ✅ Authentication views
│   ├── emails/            # ✅ Email templates
│   └── components/        # ✅ Reusable components
└── ...
```

## 📄 **Template File Extensions**

Dolphin Framework uses `.fin.html` extensions for better IDE support:

### **File Extensions**
- **`.fin.html`** - Fin template files (recommended)
- **`.fin.go`** - Legacy Fin template files
- **`.go.html`** - Go template files

### **Why `.fin.html`?**
1. **Better IDE Support**: Editors recognize HTML syntax highlighting
2. **Clear File Type**: Immediately obvious it's a template file
3. **Autocomplete**: Better autocomplete and IntelliSense
4. **Linting**: HTML/CSS linters work properly

## 🚀 **Authentication Scaffolding**

When you run `dolphin make:auth`, it creates a complete Laravel Breeze-like authentication system:

### **Generated Auth Structure**
```
views/
├── layouts/
│   └── app.fin.html           # Main application layout
├── auth/
│   ├── login.fin.html         # Login page
│   ├── register.fin.html      # Registration page
│   ├── forgot-password.fin.html # Password reset request
│   ├── reset-password.fin.html  # Password reset form
│   ├── verify-email.fin.html  # Email verification
│   └── profile.fin.html       # User profile
├── emails/
│   ├── verify-email.fin.html  # Email verification template
│   └── password-reset.fin.html # Password reset email
└── dashboard.fin.html          # User dashboard
```

### **Generated Go Files**
```
app/
├── models/
│   ├── user.go                # User model
│   └── password_reset.go      # Password reset model
├── http/controllers/
│   └── auth_controller.go     # Authentication controller
├── http/middleware/
│   └── auth.go                # Authentication middleware
├── http/requests/
│   ├── login_request.go       # Login validation
│   ├── register_request.go    # Registration validation
│   ├── forgot_password_request.go
│   ├── reset_password_request.go
│   └── update_profile_request.go
└── services/
    ├── auth_service.go        # Authentication service
    ├── mail_service.go        # Email service
    └── session_service.go     # Session service
```

## 🎨 **Tailwind CSS Integration**

Dolphin Framework includes local Tailwind CSS setup (no CDN dependency):

### **Tailwind Configuration**
```javascript
// tailwind.config.js
module.exports = {
  content: [
    "./views/**/*.{fin,fin.html,html,go}",  // Scans views directory
    "./public/**/*.html",
    "./internal/**/*.go",
  ],
  theme: {
    extend: {
      colors: {
        'dolphin-primary': '#009688',    // Teal 600
        'dolphin-secondary': '#00BCD4',  // Cyan 500
        'dolphin-light': '#e0f7fa',      // Light teal
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
```

### **CSS Build Process**
```bash
# Install dependencies
npm install

# Build CSS for development (with watch)
npm run build

# Build CSS for production (minified)
npm run build:prod
```

## 🔧 **Configuration Files**

### **Template Configuration**
```yaml
# config/template.yaml
template:
  layouts_dir: "views/layouts"
  partials_dir: "views/partials"
  pages_dir: "views/pages"
  components_dir: "views/components"
  emails_dir: "views/emails"
  extension: ".fin.html"
  auto_reload: true
  cache_templates: true
  default_layout: "app"
  enable_helpers: true
  escape_html: true
```

### **Authentication Configuration**
```yaml
# config/auth.yaml
auth:
  default_guard: "web"
  guards:
    web:
      driver: "session"
      provider: "users"
  providers:
    users:
      driver: "database"
      model: "User"
  password_reset:
    expire_minutes: 60
  email_verification:
    expire_minutes: 1440  # 24 hours
```

## 🚀 **Getting Started**

### **1. Create New Project**
```bash
dolphin new my-app
cd my-app
```

### **2. Add Authentication**
```bash
dolphin make:auth
```

### **3. Install Dependencies**
```bash
# Go dependencies
go mod tidy
go get golang.org/x/crypto/bcrypt

# Node.js dependencies
npm install
```

### **4. Build Assets**
```bash
npm run build
```

### **5. Run Migrations**
```bash
dolphin migrate
```

### **6. Start Development Server**
```bash
dolphin serve
```

## 🎯 **Benefits of New Structure**

### **1. Simplicity**
- Single `/views` directory instead of nested `/resources/views` and `/ui/views`
- Clear, intuitive organization
- Easy to find and manage templates

### **2. Better IDE Support**
- `.fin.html` extensions provide proper syntax highlighting
- Better autocomplete and IntelliSense
- HTML/CSS linters work correctly

### **3. Consistency**
- Both `dolphin new` and `dolphin make:auth` use the same structure
- No confusion about where files should go
- Predictable project layout

### **4. Modern Development**
- Local Tailwind CSS (no CDN dependency)
- Hot reloading for development
- Production-ready build process

### **5. Enterprise Ready**
- Complete authentication system
- Proper separation of concerns
- Scalable architecture

## 📚 **Next Steps**

1. **Explore Templates**: Check out the generated `.fin.html` files
2. **Customize Styles**: Modify the Tailwind configuration
3. **Add Components**: Create reusable components in `/views/components/`
4. **Extend Auth**: Add custom authentication features
5. **Deploy**: Use the production build process

---

**Dolphin Framework** - Where simplicity meets power! 🐬✨

*The new project structure makes Dolphin Framework more intuitive, maintainable, and developer-friendly than ever before.*
