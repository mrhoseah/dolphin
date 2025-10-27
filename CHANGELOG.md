# Changelog

All notable changes to this project will be documented in this file.

The format is based on Keep a Changelog and this project adheres to Semantic Versioning.

## [Unreleased]
### Added
- Debug dashboard and CLI (`debug serve|status|gc`)
- Maintenance mode with middleware and CLI
- Static page service and templates
- Generators for modules, views, repositories, providers, Postman collection
- Event system scaffolding
- README docs and examples; basic tests for debug, maintenance, static

### Changed
- **BREAKING**: Simplified project structure - now uses single `/views` directory instead of `/resources/views` and `/ui/views`
- **BREAKING**: Template files now use `.fin.html` extension instead of `.fin` for better IDE support
- **BREAKING**: Removed Raptor dependency - now uses built-in GORM migration system
- Updated `dolphin new` command to create cleaner project structure
- Enhanced `dolphin make:auth` command with Laravel Breeze-like authentication system
- Improved Tailwind CSS integration with local build system (no CDN dependency)
- Updated all documentation to reflect new project structure

### Fixed
- Fixed duplicate `package main` declaration in generated `main.go` files
- Fixed incorrect `go.mod` replace directive paths
- Fixed missing `global.css` file in Tailwind build system
- Fixed hardcoded import paths in auth scaffolding to use dynamic module names
- Fixed template rendering calls to use proper template engine

## [v0.1.0] - 2025-10-16
### Added
- Initial public release of Dolphin framework core and CLI
