package sql

import (
	"github.com/tomdionysus/trinity/schema"
)

type TableReference struct {
	Table *schema.Table
	Alias *string
}

func (me *TableReference) ToSQL(wrap bool) string {
	out := "`" + me.Table.Name + "`"
	if me.Alias != nil {
		out += " AS " + *me.Alias
	}
	return out
}
