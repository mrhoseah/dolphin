package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mrhoseah/dolphin/internal/config"
	raptor "github.com/mrhoseah/raptor/core"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Manager handles database connections and operations
type Manager struct {
	config *config.DatabaseConfig
	db     *gorm.DB
	sqlDB  *sql.DB
}

// New creates a new database manager
func New(cfg *config.DatabaseConfig) (*Manager, error) {
	manager := &Manager{
		config: cfg,
	}

	if err := manager.connect(); err != nil {
		return nil, err
	}

	return manager, nil
}

// connect establishes database connection
func (m *Manager) connect() error {
	var dialector gorm.Dialector

	switch m.config.Driver {
	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			m.config.Host, m.config.Port, m.config.Username, m.config.Password,
			m.config.Database, m.config.SSLMode)
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
			m.config.Username, m.config.Password, m.config.Host, m.config.Port,
			m.config.Database, m.config.Charset)
		dialector = mysql.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(m.config.Database)
	default:
		return fmt.Errorf("unsupported database driver: %s", m.config.Driver)
	}

	var err error
	m.db, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	// Get underlying sql.DB for connection pool configuration
	m.sqlDB, err = m.db.DB()
	if err != nil {
		return err
	}

	// Configure connection pool
	m.sqlDB.SetMaxOpenConns(m.config.MaxOpen)
	m.sqlDB.SetMaxIdleConns(m.config.MaxIdle)
	m.sqlDB.SetConnMaxLifetime(time.Duration(m.config.MaxLife) * time.Second)

	return nil
}

// GetDB returns the GORM database instance
func (m *Manager) GetDB() *gorm.DB {
	return m.db
}

// GetSQLDB returns the underlying sql.DB instance
func (m *Manager) GetSQLDB() *sql.DB {
	return m.sqlDB
}

// Close closes the database connection
func (m *Manager) Close() error {
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

// Migrator handles database migrations using Raptor
type Migrator struct {
	db            *sql.DB
	migrationsDir string
	schema        raptor.Schema
}

// MigrationResult represents the result of a migration operation
type MigrationResult struct {
	Message    string
	Executed   []string
	RolledBack []string
	Batch      int
}

// MigrationStatus represents the status of a migration
type MigrationStatus struct {
	Migration string
	Status    string
	Batch     *int
}

// NewMigrator creates a new migration manager
func NewMigrator(db *sql.DB, migrationsDir string) *Migrator {
	return &Migrator{
		db:            db,
		migrationsDir: migrationsDir,
		schema:        NewSchema(db),
	}
}

// NewSchema creates a new schema instance based on database driver
func NewSchema(db *sql.DB) raptor.Schema {
	// For now, return a generic schema
	// In a real implementation, you would determine the driver type
	return &GenericSchema{DB: db}
}

// Migrate runs all pending migrations
func (m *Migrator) Migrate() MigrationResult {
	// Get all migrations from the migrations directory
	migrations := m.getMigrations()

	// Get executed migrations
	executed := m.getExecutedMigrations()

	// Find pending migrations
	var pending []raptor.Migration
	for _, migration := range migrations {
		if !m.isExecuted(migration.Name(), executed) {
			pending = append(pending, migration)
		}
	}

	if len(pending) == 0 {
		return MigrationResult{Message: "No pending migrations"}
	}

	// Get next batch number
	batch := m.getNextBatchNumber()

	// Execute pending migrations
	var executedNames []string
	for _, migration := range pending {
		if err := migration.Up(m.schema); err != nil {
			return MigrationResult{Message: fmt.Sprintf("Migration failed: %s", err.Error())}
		}

		// Record migration
		m.recordMigration(migration.Name(), batch)
		executedNames = append(executedNames, migration.Name())
	}

	return MigrationResult{
		Message:  "Migrations completed successfully",
		Executed: executedNames,
		Batch:    batch,
	}
}

// Rollback rolls back the last batch of migrations
func (m *Migrator) Rollback() MigrationResult {
	// Get last batch
	lastBatch := m.getLastBatchNumber()
	if lastBatch == 0 {
		return MigrationResult{Message: "No migrations to rollback"}
	}

	// Get migrations in last batch
	migrations := m.getMigrationsByBatch(lastBatch)
	if len(migrations) == 0 {
		return MigrationResult{Message: "No migrations to rollback"}
	}

	// Rollback migrations in reverse order
	var rolledBack []string
	for i := len(migrations) - 1; i >= 0; i-- {
		migrationName := migrations[i]
		migration := m.findMigration(migrationName)
		if migration == nil {
			continue
		}

		if err := migration.Down(m.schema); err != nil {
			return MigrationResult{Message: fmt.Sprintf("Rollback failed: %s", err.Error())}
		}

		// Remove migration record
		m.removeMigration(migrationName)
		rolledBack = append(rolledBack, migrationName)
	}

	return MigrationResult{
		Message:    "Rollback completed successfully",
		RolledBack: rolledBack,
		Batch:      lastBatch,
	}
}

// Status returns the status of all migrations
func (m *Migrator) Status() []MigrationStatus {
	allMigrations := m.getAllMigrationNames()
	executed := m.getExecutedMigrations()

	var status []MigrationStatus
	for _, migration := range allMigrations {
		s := MigrationStatus{
			Migration: migration,
			Status:    "pending",
		}

		if m.isExecuted(migration, executed) {
			s.Status = "executed"
			if batch := m.getMigrationBatch(migration); batch != nil {
				s.Batch = batch
			}
		}

		status = append(status, s)
	}

	return status
}

// Helper methods for migration management

func (m *Migrator) getMigrations() []raptor.Migration {
	// This would typically scan the migrations directory
	// For now, return empty slice - implement based on your migration structure
	return []raptor.Migration{}
}

func (m *Migrator) getAllMigrationNames() []string {
	// This would typically scan the migrations directory for .go files
	// For now, return empty slice - implement based on your migration structure
	return []string{}
}

func (m *Migrator) findMigration(name string) raptor.Migration {
	// This would find and instantiate the migration by name
	// For now, return nil - implement based on your migration structure
	return nil
}

func (m *Migrator) getExecutedMigrations() []string {
	query := "SELECT migration FROM migrations ORDER BY id"
	rows, err := m.db.Query(query)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			continue
		}
		migrations = append(migrations, migration)
	}

	return migrations
}

func (m *Migrator) isExecuted(migration string, executed []string) bool {
	for _, e := range executed {
		if e == migration {
			return true
		}
	}
	return false
}

func (m *Migrator) getNextBatchNumber() int {
	query := "SELECT COALESCE(MAX(batch), 0) + 1 FROM migrations"
	var batch int
	m.db.QueryRow(query).Scan(&batch)
	return batch
}

func (m *Migrator) getLastBatchNumber() int {
	query := "SELECT COALESCE(MAX(batch), 0) FROM migrations"
	var batch int
	m.db.QueryRow(query).Scan(&batch)
	return batch
}

func (m *Migrator) getMigrationsByBatch(batch int) []string {
	query := "SELECT migration FROM migrations WHERE batch = ? ORDER BY id"
	rows, err := m.db.Query(query, batch)
	if err != nil {
		return []string{}
	}
	defer rows.Close()

	var migrations []string
	for rows.Next() {
		var migration string
		if err := rows.Scan(&migration); err != nil {
			continue
		}
		migrations = append(migrations, migration)
	}

	return migrations
}

func (m *Migrator) getMigrationBatch(migration string) *int {
	query := "SELECT batch FROM migrations WHERE migration = ?"
	var batch int
	err := m.db.QueryRow(query, migration).Scan(&batch)
	if err != nil {
		return nil
	}
	return &batch
}

func (m *Migrator) recordMigration(migration string, batch int) {
	query := "INSERT INTO migrations (migration, batch) VALUES (?, ?)"
	m.db.Exec(query, migration, batch)
}

func (m *Migrator) removeMigration(migration string) {
	query := "DELETE FROM migrations WHERE migration = ?"
	m.db.Exec(query, migration)
}
