package schema

const (
	SQLInt     = 1
	SQLChar    = 2
	SQLNChar   = 3
	SQLVarChar = 3
	SQLFloat   = 4
	SQLDouble  = 5
)

type SQLType byte

type Field struct {
	Name     string
	SQLType  SQLType
	Length   uint
	Nullable bool
	Table    *Table
}

func (fd *Field) ToSQL() string {
	return fd.Name
}
