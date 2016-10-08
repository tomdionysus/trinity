package schema

import (
	"errors"
)

var FieldExistsError = errors.New("Field Exists")
var FieldNotFoundError = errors.New("Field Not Found")

type Table struct {
	Name    string
	Fields  map[string]*Field
	Indexes []*Index
}

func NewTable(name string) *Table {
	return &Table{
		Name:   name,
		Fields: map[string]*Field{},
	}
}

func (tb *Table) AddField(field *Field) error {
	if _, exists := tb.Fields[field.Name]; exists {
		return FieldExistsError
	}
	tb.Fields[field.Name] = field
	return nil
}

func (tb *Table) RemoveField(field *Field) error {
	if _, exists := tb.Fields[field.Name]; !exists {
		return FieldNotFoundError
	}
	delete(tb.Fields, field.Name)
	return nil
}

func (tb *Table) RemoveFieldByName(fieldName string) error {
	if _, exists := tb.Fields[fieldName]; !exists {
		return FieldNotFoundError
	}
	delete(tb.Fields, fieldName)
	return nil
}

func (tb *Table) ToSQL(wrap bool) string {
	if wrap {
		return "(" + tb.Name + ")"
	}
	return tb.Name
}
