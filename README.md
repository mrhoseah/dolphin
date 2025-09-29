🐬 Dolphin Framework: Go with Laravel Grace

The Dolphin Framework is a rapid development web framework written in Go, inspired by the elegant syntax, architecture, and developer experience of Laravel.

It aims to blend Go's performance and concurrency capabilities with the productivity features of modern PHP frameworks, such as a Service Container (Kernel), an Active Record ORM (DORM), and a robust CLI (D-CLI).
🚀 Getting Started
Installation (Conceptual)

Since Dolphin is a pure Go framework, you can pull it in using standard Go tooling:

# 1. Start a new module
go mod init my-dolphin-app

# 2. Add required dependencies (Go-Chi, etc.)
# go get [github.com/go-chi/chi/v5](https://github.com/go-chi/chi/v5)
# go get gorm.io/gorm

Running the Application

Use the Dolphin Command Line Interface (D-CLI) to start the development server:

go run dolphin_cli.go serve
# Dolphin server running at http://localhost:8080

Creating Components

Use D-CLI to quickly scaffold application files:

# Create a new HTTP Controller
go run dolphin_cli.go make:controller PostController
# Creates: app/http/controllers/PostController.go

🛠️ Architecture Overview

Dolphin utilizes a clear separation of concerns, anchored by its core components:

    Kernel (pkg/core/kernel.go): The central Service Container that manages dependency injection, application bootstrapping, and the router instance.

    Routing (routes/*.go): All routes are defined separately from the Kernel, providing clear separation between framework bootstrapping and application routing logic.

    Controller (app/http/controllers): The "C" in MVC. Go structs that handle HTTP requests and rely on Dependency Injection for services and models.

    DORM (pkg/orm/model.go): The Dolphin ORM, an Active Record implementation for effortless database interaction (the "M" in MVC).

    D-CLI (dolphin_cli.go): The Command-Line Interface, essential for scaffolding, migrations, and developer utilities.