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

func (me *Comparison) ToSQL(wrap bool) string {
	out := ""
	if wrap {
		out += "("
	}
	if me.A != nil {
		out += me.A.ToSQL(false)
	}
	out += " " + me.GetOperationString() + " "
	if me.B != nil {
		out += me.B.ToSQL(false)
	}
	if wrap {
		out += ")"
	}
	return out
}

func (me *Comparison) GetOperationString() string {
	switch me.Operation {
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
