package sql

import (
	"github.com/tomdionysus/trinity/schema"
)

type FieldReference struct {
	Field  *schema.Field
	Source Term
	Alias  *string
}

func (me *FieldReference) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	if me.Source != nil {
		out += me.Source.ToSQL(false) + "."
	}
	out += me.Field.Name
	if me.Alias != nil {
		out += " AS " + *me.Alias
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
