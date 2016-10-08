package sql

type CreateDatabase struct {
	Name string
}

func (cdb *CreateDatabase) ToSQL(wrap bool) string {
	return "CREATE DATABASE `" + cdb.Name + "`"
}
