package sql

import (
	"github.com/tomdionysus/trinity/schema"
)

type FieldReference struct {
	Field  *schema.Field
	Source Term
	Alias  *string
}

func (fr *FieldReference) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	if fr.Source != nil {
		out += fr.Source.ToSQL(false) + "."
	}
	out += fr.Field.Name
	if fr.Alias != nil {
		out += " AS " + *fr.Alias
	}
	if wrap {
		out += ")"
	}
	return out
}

func NewFieldReference(field *schema.Field, source Term, alias *string) *FieldReference {
	return &FieldReference{
		Field:  field,
		Source: source,
		Alias:  alias,
	}
}
