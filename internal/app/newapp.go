package app

import (
	"fmt"
	"os"
	"path/filepath"
)

// createGoMod creates go.mod file
func (g *Generator) createGoMod(appName string) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	dolphin v0.0.0
)

replace dolphin => ../dolphin
`, appName)
	return os.WriteFile("go.mod", []byte(content), 0644)
}

// createMainGo creates main.go file
func (g *Generator) createMainGo(appName string) error {
	content := fmt.Sprintf(`package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"dolphin/internal/config"
	"dolphin/internal/database"
	"dolphin/internal/logger"
	"dolphin/internal/router"
	"dolphin/pkg/app"
	"%s/bootstrap"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize logger
	logger := logger.New(cfg.Log.Level, cfg.Log.Format)

	// Initialize database
	db, err := database.New(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Initialize application
	application := app.New(cfg, logger, db)

	// Initialize router
	r := router.New(application)

	// Setup routes
	bootstrap.SetupRoutes(r, application, logger, db.DB)
	r.SetupRoutes()

	// Create root command
	var rootCmd = &cobra.Command{
		Use:   "%s",
		Short: "Dolphin Application",
	}

	// Add serve command
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the development server",
		Run: func(cmd *cobra.Command, args []string) {
			port := cfg.Server.Port
			host := cfg.Server.Host

			logger.Info("🚀 Starting server...", zap.String("host", host), zap.Int("port", port))
			logger.Info("📚 API Documentation available at http://" + host + ":" + fmt.Sprintf("%%d", port) + "/swagger/index.html")

			if err := http.ListenAndServe(fmt.Sprintf("%%s:%%d", host, port), r); err != nil {
				logger.Fatal("Failed to start server", zap.Error(err))
			}
		},
	}

	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		logger.Fatal("Command execution failed", zap.Error(err))
	}
}
`, appName, appName)
	return os.WriteFile("main.go", []byte(content), 0644)
}

// createConfigFiles creates configuration files
func (g *Generator) createConfigFiles() error {
	// Create config.yaml
	configContent := `app:
  name: "Dolphin App"
  env: "development"
  debug: true
  url: "http://localhost:8080"

server:
  host: "localhost"
  port: 8080
  read_timeout: 30
  write_timeout: 30
  idle_timeout: 120

database:
  driver: "sqlite"
  database: "app.db"
  host: ""
  port: 0
  username: ""
  password: ""
  ssl_mode: "disable"

log:
  level: "debug"
  format: "json"
  output: "stdout"
`
	if err := os.WriteFile("config/config.yaml", []byte(configContent), 0644); err != nil {
		return err
	}

	return nil
}

// createBootstrapRoutes creates bootstrap routes file
func (g *Generator) createBootstrapRoutes() error {
	content := `package bootstrap

import (
	"net/http"
	"dolphin/internal/router"
	"dolphin/pkg/app"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func SetupRoutes(r *router.Router, application *app.Application, logger *zap.Logger, db *gorm.DB) {
	chiRouter := r.GetRouter()
	templateEngine := r.GetFinEngine()

	// Home route
	chiRouter.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if err := templateEngine.Render(w, "pages/home.fin.html", nil); err != nil {
			logger.Error("Failed to render home page", zap.Error(err))
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
}
`
	return os.WriteFile("bootstrap/routes.go", []byte(content), 0644)
}

// createReactPackageJson creates package.json for React
func (g *Generator) createReactPackageJson() error {
	content := `{
  "name": "dolphin-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "npm run build:ts:watch & npm run build:css:watch",
    "build": "npm run build:ts && npm run build:css && node scripts/build-manifest.js",
    "build:ts": "esbuild spa/ts/app.tsx --bundle --outfile=public/js/app.js --target=es2020 --format=iife --jsx=automatic --sourcemap",
    "build:ts:watch": "esbuild spa/ts/app.tsx --bundle --outfile=public/js/app.js --target=es2020 --format=iife --jsx=automatic --sourcemap --watch",
    "build:css": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --minify",
    "build:css:watch": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --watch"
  },
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0"
  },
  "devDependencies": {
    "@types/react": "^18.2.0",
    "@types/react-dom": "^18.2.0",
    "esbuild": "^0.19.0",
    "typescript": "^5.0.0",
    "tailwindcss": "^3.4.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32"
  }
}
`
	if err := os.WriteFile("package.json", []byte(content), 0644); err != nil { return err }
	// write manifest builder script
	return g.createManifestScript()
}

// createVuePackageJson creates package.json for Vue
func (g *Generator) createVuePackageJson() error {
	content := `{
  "name": "dolphin-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "npm run build:watch",
    "build": "npm run build:js && npm run build:css && node scripts/build-manifest.js",
    "build:js": "esbuild spa/js/app.js --bundle --outfile=public/js/app.js --minify --sourcemap",
    "build:watch": "esbuild spa/js/app.js --bundle --outfile=public/js/app.js --sourcemap --watch",
    "build:css": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --minify",
    "build:css:watch": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --watch"
  },
  "dependencies": {
    "vue": "^3.4.0"
  },
  "devDependencies": {
    "esbuild": "^0.19.0",
    "tailwindcss": "^3.4.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32"
  }
}
`
	if err := os.WriteFile("package.json", []byte(content), 0644); err != nil { return err }
	return g.createManifestScript()
}

func (g *Generator) createManifestScript() error {
	if err := os.MkdirAll("scripts", 0755); err != nil { return err }
	script := `const fs=require('fs');const path=require('path');
function hash(content){const crypto=require('crypto');return crypto.createHash('sha1').update(content).digest('hex').slice(0,8)}
function processFile(rel){const p=path.join('public',rel);if(!fs.existsSync(p))return null;const c=fs.readFileSync(p);const h=hash(c);const ext=path.extname(rel);const base=rel.slice(0,-ext.length);const out=base+'.'+h+ext;fs.copyFileSync(p,path.join('public',out));return [rel,out];}
const manifest={};['js/app.js','css/app.css'].forEach(f=>{const r=processFile(f);if(r) manifest[r[0]]=r[1];});fs.writeFileSync(path.join('public','manifest.json'),JSON.stringify(manifest, null, 2));
`
	return os.WriteFile("scripts/build-manifest.js", []byte(script), 0644)
}

// createFinPackageJson creates package.json for Fin templates
func (g *Generator) createFinPackageJson() error {
	content := `{
  "name": "dolphin-app",
  "version": "1.0.0",
  "scripts": {
    "dev": "npm run build:watch",
    "build": "npm run build:js && npm run build:css",
    "build:js": "esbuild assets/js/app.js --bundle --outfile=public/js/app.js --minify --sourcemap",
    "build:watch": "esbuild assets/js/app.js --bundle --outfile=public/js/app.js --sourcemap --watch",
    "build:css": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --minify",
    "build:css:watch": "tailwindcss -i ./assets/css/app.css -o ./public/css/app.css --watch"
  },
  "devDependencies": {
    "esbuild": "^0.19.0",
    "tailwindcss": "^3.4.0",
    "autoprefixer": "^10.4.16",
    "postcss": "^8.4.32"
  }
}
`
	return os.WriteFile("package.json", []byte(content), 0644)
}

// createTsConfig creates tsconfig.json
func (g *Generator) createTsConfig() error {
	content := `{
  "compilerOptions": {
    "target": "ES2020",
    "module": "ES2020",
    "lib": ["ES2020", "DOM"],
    "jsx": "react",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true
  }
}
`
	return os.WriteFile("tsconfig.json", []byte(content), 0644)
}

// createTailwindConfig creates tailwind.config.js
func (g *Generator) createTailwindConfig() error {
	content := `/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    './views/**/*.fin.html',
    './spa/**/*.{js,ts,jsx,tsx}',
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
`
	return os.WriteFile("tailwind.config.js", []byte(content), 0644)
}

// createPostCSSConfig creates postcss.config.js
func (g *Generator) createPostCSSConfig() error {
	content := `module.exports = {
  plugins: {
    tailwindcss: {},
    autoprefixer: {},
  },
}
`
	return os.WriteFile("postcss.config.js", []byte(content), 0644)
}

// createReactEntryPoint creates React entry point
func (g *Generator) createReactEntryPoint() error {
	if err := os.MkdirAll("spa/ts", 0755); err != nil { return err }
	content := `import React from 'react'
import { createRoot } from 'react-dom/client'
function App(){return(<div className="container mx-auto px-4 py-8"><h1 className="text-4xl font-bold">React App</h1></div>)}
const tag=document.getElementById('react-login')||document.getElementById('app');if(tag){const root=createRoot(tag);root.render(<App/>)}
`
	return os.WriteFile("spa/ts/app.tsx", []byte(content), 0644)
}

// createVueEntryPoint creates Vue entry point
func (g *Generator) createVueEntryPoint() error {
	if err := os.MkdirAll("spa/js", 0755); err != nil { return err }
	content := `import { createApp } from 'vue'
const App={template:` + "`" + `<div class=\"container mx-auto px-4 py-8\"><h1 class=\"text-4xl font-bold\">Vue App</h1></div>` + "`" + `}
const el=document.getElementById('vue-login')?'#vue-login':'#app';createApp(App).mount(el)
`
	return os.WriteFile("spa/js/app.js", []byte(content), 0644)
}

// createVanillaJSEntryPoint creates vanilla JS entry point
func (g *Generator) createVanillaJSEntryPoint() error {
	content := `// Dolphin App - Vanilla JavaScript
console.log('Dolphin app loaded')
`
	return os.WriteFile("assets/js/app.js", []byte(content), 0644)
}

// createAppCSS creates app.css
func (g *Generator) createAppCSS() error {
	content := `@tailwind base;
@tailwind components;
@tailwind utilities;
`
	dir := "assets/css"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "app.css"), []byte(content), 0644)
}

// createBaseLayoutForNewApp creates base layout template for new app
func (g *Generator) createBaseLayoutForNewApp(frontend string) error {
	var content string
	switch frontend {
	case "react":
		content = g.generateReactLayout()
	case "vue":
		content = g.generateVueLayout()
	default:
		content = g.generateFinLayout()
	}
	return os.WriteFile("views/layouts/base.fin.html", []byte(content), 0644)
}

// createHomePageForNewApp creates home page for new app
func (g *Generator) createHomePageForNewApp(frontend string) error {
	content := `{{extend "layouts/base.fin.html"}}
{{define "title"}}Home{{end}}
{{define "content"}}
<div class="container mx-auto px-4 py-8">
    <h1 class="text-4xl font-bold mb-4">Welcome to Dolphin</h1>
    <p class="text-lg text-gray-600">Your application is ready to go!</p>
</div>
{{end}}
`
	return os.WriteFile("views/pages/home.fin.html", []byte(content), 0644)
}

// generateReactLayout generates React layout template
func (g *Generator) generateReactLayout() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Dolphin App{{end}}</title>
    <link rel="stylesheet" href="/css/app.css">
</head>
<body>
    <div id="app">{{block "content" .}}{{end}}</div>
    <script id="app-script" src="/js/app.js"></script>
    <script>
    fetch('/manifest.json').then(r=>r.ok?r.json():null).then(m=>{if(m&&m['js/app.js']){var s=document.getElementById('app-script');if(s){s.src='/' + m['js/app.js'];}}}).catch(()=>{})
    </script>
</body>
</html>
`
}

// generateVueLayout generates Vue layout template
func (g *Generator) generateVueLayout() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Dolphin App{{end}}</title>
    <link rel="stylesheet" href="/css/app.css">
</head>
<body>
    <div id="app">{{block "content" .}}{{end}}</div>
    <script id="app-script" src="/js/app.js"></script>
    <script>
    fetch('/manifest.json').then(r=>r.ok?r.json():null).then(m=>{if(m&&m['js/app.js']){var s=document.getElementById('app-script');if(s){s.src='/' + m['js/app.js'];}}}).catch(()=>{})
    </script>
</body>
</html>
`
}

// generateFinLayout generates Fin layout template
func (g *Generator) generateFinLayout() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{block "title" .}}Dolphin App{{end}}</title>
    <link rel="stylesheet" href="/css/app.css">
    <script src="https://unpkg.com/htmx.org@1.9.10"></script>
</head>
<body class="bg-gray-50">
    <main>{{block "content" .}}{{end}}</main>
    <script src="/js/app.js"></script>
</body>
</html>
`
}
