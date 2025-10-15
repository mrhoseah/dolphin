package migrations

import (
	raptor "github.com/mrhoseah/raptor/core"
)

// create_products_table represents the create_products_table migration
type create_products_table struct{}

// Name returns the migration name
func (m *create_products_table) Name() string {
	return "create_products_table"
}

// Up runs the migration
func (m *create_products_table) Up(s raptor.Schema) error {
	// Add your migration logic here
	// Example: Create a table
	// return s.CreateTable("create_products_table", []string{"id", "name", "email", "created_at"})
	
	return nil
}

// Down rolls back the migration
func (m *create_products_table) Down(s raptor.Schema) error {
	// Add your rollback logic here
	// Example: Drop a table
	// return s.DropTable("create_products_table")
	
	return nil
}
