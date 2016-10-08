package sql

import (
	"strings"
)

type Select struct {
	Results []Term
	Sources []Term
	Where   Term
	Order   Term
	Limit   Term
}

func (me *Select) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	out += "SELECT "
	if len(me.Results) > 0 {
		out += GetStringWithSeperator(me.Results, ", ", false)
	}
	if len(me.Sources) > 0 {
		out += " FROM "
		out += GetStringWithSeperator(me.Sources, ", ", false)
	}
	if me.Where != nil {
		out += " WHERE "
		out += me.Where.ToSQL(false)
	}
	if me.Order != nil {
		out += " ORDER BY "
		out += me.Order.ToSQL(false)
	}
	if me.Limit != nil {
		out += " LIMIT "
		out += me.Limit.ToSQL(false)
	}
	if wrap {
		out += ")"
	}
	return out
}

func GetStringWithSeperator(terms []Term, separator string, wrap bool) string {
	out := []string{}
	for _, term := range terms {
		out = append(out, term.ToSQL(wrap))
	}
	return strings.Join(out, separator)
}
