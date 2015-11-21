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

func (me *Table) AddField(field *Field) error {
	if _, exists := me.Fields[field.Name]; exists {
		return FieldExistsError
	}
	me.Fields[field.Name] = field
	return nil
}

func (me *Table) RemoveField(field *Field) error {
	if _, exists := me.Fields[field.Name]; !exists {
		return FieldNotFoundError
	}
	delete(me.Fields, field.Name)
	return nil
}

func (me *Table) RemoveFieldByName(fieldName string) error {
	if _, exists := me.Fields[fieldName]; !exists {
		return FieldNotFoundError
	}
	delete(me.Fields, fieldName)
	return nil
}

func (me *Table) ToSQL(wrap bool) string {
	if wrap {
		return "(" + me.Name + ")"
	}
	return me.Name
}
