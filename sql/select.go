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

func (sel *Select) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	out += "SELECT "
	if len(sel.Results) > 0 {
		out += GetStringWithSeperator(sel.Results, ", ", false)
	}
	if len(sel.Sources) > 0 {
		out += " FROM "
		out += GetStringWithSeperator(sel.Sources, ", ", false)
	}
	if sel.Where != nil {
		out += " WHERE "
		out += sel.Where.ToSQL(false)
	}
	if sel.Order != nil {
		out += " ORDER BY "
		out += sel.Order.ToSQL(false)
	}
	if sel.Limit != nil {
		out += " LIMIT "
		out += sel.Limit.ToSQL(false)
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
