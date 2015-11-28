package sql

type CreateDatabase struct {
	Name string
}

func (me *CreateDatabase) ToSQL(wrap bool) string {
	return "CREATE DATABASE `" + me.Name + "`"
}
