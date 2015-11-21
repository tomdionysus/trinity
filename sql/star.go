package sql

type Star struct {
}

func (me *Star) ToSQL(wrap bool) string {
	if wrap {
		return "(*)"
	}
	return "*"
}
