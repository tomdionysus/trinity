package sql

type Star struct {
}

func (star *Star) ToSQL(wrap bool) string {
	if wrap {
		return "(*)"
	}
	return "*"
}
