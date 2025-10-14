package database

import (
	"database/sql"
	"fmt"
	"strings"

	raptor "github.com/mrhoseah/raptor/core"
)

// PostgresSchema implements raptor.Schema for PostgreSQL
type PostgresSchema struct {
	DB *sql.DB
}

var _ raptor.Schema = (*PostgresSchema)(nil)

func (s *PostgresSchema) CreateTable(name string, columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	// Build column definitions
	var columnDefs []string
	for _, col := range columns {
		switch col {
		case "id":
			columnDefs = append(columnDefs, "id SERIAL PRIMARY KEY")
		case "email":
			columnDefs = append(columnDefs, "email VARCHAR(255) UNIQUE NOT NULL")
		case "password":
			columnDefs = append(columnDefs, "password VARCHAR(255) NOT NULL")
		case "created_at":
			columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		case "updated_at":
			columnDefs = append(columnDefs, "updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		default:
			columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(255)", col))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s)", name, strings.Join(columnDefs, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) DropTable(name string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) AddColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) DropColumn(table, column string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) RenameColumn(table, oldName, newName string) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldName, newName)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) ChangeColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) AddIndex(table, name string, columns []string) error {
	query := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", name, table, strings.Join(columns, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) DropIndex(table, name string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) AddForeignKey(table, name, column, refTable, refColumn string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		table, name, column, refTable, refColumn)
	_, err := s.DB.Exec(query)
	return err
}

func (s *PostgresSchema) DropForeignKey(table, name string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", table, name)
	_, err := s.DB.Exec(query)
	return err
}

// MySQLSchema implements raptor.Schema for MySQL
type MySQLSchema struct {
	DB *sql.DB
}

var _ raptor.Schema = (*MySQLSchema)(nil)

func (s *MySQLSchema) CreateTable(name string, columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	// Build column definitions
	var columnDefs []string
	for _, col := range columns {
		switch col {
		case "id":
			columnDefs = append(columnDefs, "id INT AUTO_INCREMENT PRIMARY KEY")
		case "email":
			columnDefs = append(columnDefs, "email VARCHAR(255) UNIQUE NOT NULL")
		case "password":
			columnDefs = append(columnDefs, "password VARCHAR(255) NOT NULL")
		case "created_at":
			columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		case "updated_at":
			columnDefs = append(columnDefs, "updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP")
		default:
			columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(255)", col))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4", name, strings.Join(columnDefs, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) DropTable(name string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) AddColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) DropColumn(table, column string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) RenameColumn(table, oldName, newName string) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldName, newName)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) ChangeColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s MODIFY COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) AddIndex(table, name string, columns []string) error {
	query := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", name, table, strings.Join(columns, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) DropIndex(table, name string) error {
	query := fmt.Sprintf("DROP INDEX %s ON %s", name, table)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) AddForeignKey(table, name, column, refTable, refColumn string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		table, name, column, refTable, refColumn)
	_, err := s.DB.Exec(query)
	return err
}

func (s *MySQLSchema) DropForeignKey(table, name string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP FOREIGN KEY %s", table, name)
	_, err := s.DB.Exec(query)
	return err
}

// SQLiteSchema implements raptor.Schema for SQLite
type SQLiteSchema struct {
	DB *sql.DB
}

var _ raptor.Schema = (*SQLiteSchema)(nil)

func (s *SQLiteSchema) CreateTable(name string, columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	// Build column definitions
	var columnDefs []string
	for _, col := range columns {
		switch col {
		case "id":
			columnDefs = append(columnDefs, "id INTEGER PRIMARY KEY AUTOINCREMENT")
		case "email":
			columnDefs = append(columnDefs, "email TEXT UNIQUE NOT NULL")
		case "password":
			columnDefs = append(columnDefs, "password TEXT NOT NULL")
		case "created_at":
			columnDefs = append(columnDefs, "created_at DATETIME DEFAULT CURRENT_TIMESTAMP")
		case "updated_at":
			columnDefs = append(columnDefs, "updated_at DATETIME DEFAULT CURRENT_TIMESTAMP")
		default:
			columnDefs = append(columnDefs, fmt.Sprintf("%s TEXT", col))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s)", name, strings.Join(columnDefs, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *SQLiteSchema) DropTable(name string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *SQLiteSchema) AddColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *SQLiteSchema) DropColumn(table, column string) error {
	// SQLite doesn't support DROP COLUMN directly
	// This would require recreating the table
	return fmt.Errorf("SQLite does not support DROP COLUMN directly")
}

func (s *SQLiteSchema) RenameColumn(table, oldName, newName string) error {
	// SQLite doesn't support RENAME COLUMN in older versions
	return fmt.Errorf("SQLite does not support RENAME COLUMN")
}

func (s *SQLiteSchema) ChangeColumn(table, column, definition string) error {
	// SQLite doesn't support MODIFY COLUMN
	return fmt.Errorf("SQLite does not support MODIFY COLUMN")
}

func (s *SQLiteSchema) AddIndex(table, name string, columns []string) error {
	query := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", name, table, strings.Join(columns, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *SQLiteSchema) DropIndex(table, name string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *SQLiteSchema) AddForeignKey(table, name, column, refTable, refColumn string) error {
	// SQLite foreign key support is limited
	return fmt.Errorf("SQLite foreign key support is limited")
}

func (s *SQLiteSchema) DropForeignKey(table, name string) error {
	// SQLite foreign key support is limited
	return fmt.Errorf("SQLite foreign key support is limited")
}

// GenericSchema implements raptor.Schema for generic SQL databases
type GenericSchema struct {
	DB *sql.DB
}

var _ raptor.Schema = (*GenericSchema)(nil)

func (s *GenericSchema) CreateTable(name string, columns []string) error {
	if len(columns) == 0 {
		return fmt.Errorf("at least one column is required")
	}

	// Build column definitions
	var columnDefs []string
	for _, col := range columns {
		switch col {
		case "id":
			columnDefs = append(columnDefs, "id INTEGER PRIMARY KEY")
		case "email":
			columnDefs = append(columnDefs, "email VARCHAR(255) UNIQUE NOT NULL")
		case "password":
			columnDefs = append(columnDefs, "password VARCHAR(255) NOT NULL")
		case "created_at":
			columnDefs = append(columnDefs, "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		case "updated_at":
			columnDefs = append(columnDefs, "updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
		default:
			columnDefs = append(columnDefs, fmt.Sprintf("%s VARCHAR(255)", col))
		}
	}

	query := fmt.Sprintf("CREATE TABLE %s (%s)", name, strings.Join(columnDefs, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) DropTable(name string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) AddColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) DropColumn(table, column string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s", table, column)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) RenameColumn(table, oldName, newName string) error {
	query := fmt.Sprintf("ALTER TABLE %s RENAME COLUMN %s TO %s", table, oldName, newName)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) ChangeColumn(table, column, definition string) error {
	query := fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s %s", table, column, definition)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) AddIndex(table, name string, columns []string) error {
	query := fmt.Sprintf("CREATE INDEX %s ON %s (%s)", name, table, strings.Join(columns, ", "))
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) DropIndex(table, name string) error {
	query := fmt.Sprintf("DROP INDEX IF EXISTS %s", name)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) AddForeignKey(table, name, column, refTable, refColumn string) error {
	query := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)",
		table, name, column, refTable, refColumn)
	_, err := s.DB.Exec(query)
	return err
}

func (s *GenericSchema) DropForeignKey(table, name string) error {
	query := fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT %s", table, name)
	_, err := s.DB.Exec(query)
	return err
}
