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

func (me *Database) AddTable(table *Table) error {
	if _, exists := me.Tables[table.Name]; exists {
		return TableExistsError
	}
	me.Tables[table.Name] = table
	return nil
}

func (me *Database) RemoveTable(table *Table) error {
	if _, exists := me.Tables[table.Name]; !exists {
		return TableNotFoundError
	}
	delete(me.Tables, table.Name)
	return nil
}

func (me *Database) RemoveTableByName(tableName string) error {
	if _, exists := me.Tables[tableName]; !exists {
		return TableNotFoundError
	}
	delete(me.Tables, tableName)
	return nil
}
