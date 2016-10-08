package sql

type SQLOperation byte

const (
	OperationAND = iota
	OperationOR  = iota
	OperationNOT = iota
	OperationEQ  = iota
	OperationNEQ = iota
	OperationLT  = iota
	OperationGT  = iota
	OperationLTE = iota
	OperationGTE = iota
)

type Comparison struct {
	A         Term
	B         Term
	Operation SQLOperation
}

func (cmp *Comparison) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	if cmp.A != nil {
		out += cmp.A.ToSQL(false)
	}
	out += " " + cmp.GetOperationString() + " "
	if cmp.B != nil {
		out += cmp.B.ToSQL(false)
	}
	if wrap {
		out += ")"
	}
	return out
}

func (cmp *Comparison) GetOperationString() string {
	switch cmp.Operation {
	case OperationAND:
		return "AND"
	case OperationOR:
		return "OR"
	case OperationNOT:
		return "NOT"
	case OperationEQ:
		return "="
	case OperationNEQ:
		return "<>"
	case OperationLT:
		return "<"
	case OperationGT:
		return ">"
	case OperationLTE:
		return "<="
	case OperationGTE:
		return ">="
	default:
		return "!UNKNOWN!"
	}
}
