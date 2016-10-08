package sql

import (
	"github.com/tomdionysus/trinity/schema"
)

type Constant struct {
	SQLType schema.SQLType
	Value   string
}

func (cst *Constant) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	switch cst.SQLType {
	case schema.SQLVarChar:
		out += "\"" + cst.Value + "\""
	default:
		out += cst.Value
	}
	if wrap {
		out += ")"
	}
	return out
}

func NewConstant(sqlType schema.SQLType, value string) *Constant {
	return &Constant{
		SQLType: sqlType,
		Value:   value,
	}
}
