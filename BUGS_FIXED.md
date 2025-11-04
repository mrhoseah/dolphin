# Bugs Fixed in Dolphin Framework

## Bug #1: Route Conflict When Custom Routes Are Registered
**Severity**: High  
**Status**: ✅ Fixed

### Description
When apps register custom routes (e.g., `/auth`) and then call `SetupRoutes()`, Dolphin's default routes attempted to register on the same paths, causing a panic:
```
panic: chi: attempting to Mount() a handler on an existing path, '/auth'
```

### Root Cause
The `setupWebRoutes()` function in `internal/router/web.go` was registering default routes (home, auth, dashboard, admin, partials) without checking if custom routes had already been registered.

### Fix
Modified `setupWebRoutes()` to skip default route registration when apps provide their own routes. Apps should register custom routes before calling `SetupRoutes()`.

**Files Changed**:
- `internal/router/web.go` - Removed conflicting default route registrations

---

## Bug #2: Router Auto-Setup Breaking Custom Route Priority
**Severity**: Medium  
**Status**: ✅ Fixed

### Description
The router was automatically setting up routes in `New()`, which meant custom routes registered afterwards would not take precedence over default routes.

### Root Cause
Chi router matches routes in registration order. If default routes are registered first, they take precedence.

### Fix
Changed router initialization to not automatically set up routes. Apps must now:
1. Create router: `r := router.New(app)`
2. Register custom routes: `bootstrap.SetupRoutes(r, ...)`
3. Setup default routes: `r.SetupRoutes()` (custom routes will take precedence)

**Files Changed**:
- `internal/router/router.go` - Removed automatic `setupRoutes()` call
- `pkg/router/router.go` - Added public `SetupRoutes()` method

---

## Bug #3: Template Engine Not Accessible from Router
**Severity**: Medium  
**Status**: ✅ Fixed

### Description
Apps couldn't access the router's template engine to render views in custom routes, forcing them to create a new template engine instance with potentially different configuration.

### Root Cause
The `FinTemplateEngine` was stored in the internal router struct but not exposed via the public API.

### Fix
Added `GetFinEngine()` method to both internal and public router APIs, allowing apps to access the shared template engine instance.

**Files Changed**:
- `internal/router/router.go` - Added `GetFinEngine()` method
- `pkg/router/router.go` - Added `GetFinEngine()` wrapper method

---

## Bug #4: Generator Not Creating Essential Pages
**Severity**: Low  
**Status**: ✅ Fixed

### Description
The `make:auth` command was not generating essential pages like `home.fin.html` and `dashboard.fin.html`, causing "view not found" errors.

### Root Cause
The `CreateAuth()` function in the generator only created authentication views, not the base pages needed for a complete app.

### Fix
Enhanced `CreateAuth()` to also generate:
- `views/pages/home.fin.html` - Home page with welcome message
- `views/pages/dashboard.fin.html` - Dashboard page with stats and quick actions

**Files Changed**:
- `internal/app/generator.go` - Added `createHomePage()` and `createDashboardPage()` methods
- `internal/app/templates.go` - Added `generateHomePageContent()` and `generateDashboardPageContent()` methods

---

## Summary of Changes

### Modified Files:
1. `internal/router/router.go`
   - Removed automatic route setup from `New()`
   - Added `SetupRoutes()` public method
   - Added `GetFinEngine()` method

2. `internal/router/web.go`
   - Modified `setupWebRoutes()` to skip default routes when apps provide custom ones

3. `pkg/router/router.go`
   - Added `GetFinEngine()` wrapper
   - Added `SetupRoutes()` wrapper

4. `internal/app/generator.go`
   - Enhanced `CreateAuth()` to generate home and dashboard pages

5. `internal/app/templates.go`
   - Added templates for home and dashboard pages

### Breaking Changes:
- Apps must now call `r.SetupRoutes()` after registering custom routes
- Default web routes (home, auth, dashboard, admin) are no longer automatically registered

### Migration Guide:
**Before:**
```go
r := router.New(app)
// Custom routes...
```

**After:**
```go
r := router.New(app)
// Register custom routes first
bootstrap.SetupRoutes(r, ...)
// Then setup default routes
r.SetupRoutes()
```

---

## Testing Recommendations

1. Test route precedence: Custom routes should take precedence over defaults
2. Test template engine access: Verify `GetFinEngine()` returns the correct instance
3. Test generator: Verify `make:auth` creates all expected files
4. Test backward compatibility: Ensure apps without custom routes still work

