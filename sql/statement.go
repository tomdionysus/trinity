package sql

const (
	// Commands
	CreateDatabase = 1
	DropDatabase   = 2

	CreateTable = 3
	DropTable   = 4

	SQLInsert = 5
	SQLSelect = 6
	SQLUpdate = 7
	SQLDelete = 8
)

type Term interface {
	ToSQL(wrap bool) string
}

type StatementCommand int
type DataType byte
type TermOperation byte

type Where struct {
	Terms []*Term
}
