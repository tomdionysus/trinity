package sql

import (
	"github.com/tomdionysus/trinity/schema"
)

type TableReference struct {
	Table *schema.Table
	Alias *string
}

func (tblref *TableReference) ToSQL(wrap bool) string {
	out := "`" + tblref.Table.Name + "`"
	if tblref.Alias != nil {
		out += " AS " + *tblref.Alias
	}
	return out
}
