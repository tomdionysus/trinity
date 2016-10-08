package schema

import (
	"errors"
)

var TableExistsError = errors.New("Table Exists")
var TableNotFoundError = errors.New("Table Not Found")

type Database struct {
	Name   string
	Tables map[string]*Table
}

func NewDatabase(name string) *Database {
	inst := &Database{
		Name:   name,
		Tables: map[string]*Table{},
	}
	return inst
}

func (db *Database) AddTable(table *Table) error {
	if _, exists := db.Tables[table.Name]; exists {
		return TableExistsError
	}
	db.Tables[table.Name] = table
	return nil
}

func (db *Database) RemoveTable(table *Table) error {
	if _, exists := db.Tables[table.Name]; !exists {
		return TableNotFoundError
	}
	delete(db.Tables, table.Name)
	return nil
}

func (db *Database) RemoveTableByName(tableName string) error {
	if _, exists := db.Tables[tableName]; !exists {
		return TableNotFoundError
	}
	delete(db.Tables, tableName)
	return nil
}
