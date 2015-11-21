package main

import (
	"fmt"
	"github.com/tomdionysus/trinity/schema"
	"github.com/tomdionysus/trinity/sql"
)

func main() {

	fmt.Printf("Trinity v%s\n", VERSION)

	db := schema.NewDatabase("geodb")

	nameField := &schema.Field{Name: "name", SQLType: schema.SQLVarChar, Length: 32, Nullable: false}
	countryField := &schema.Field{Name: "country", SQLType: schema.SQLChar, Length: 2, Nullable: false}
	populationField := &schema.Field{Name: "country", SQLType: schema.SQLChar, Length: 2, Nullable: false}

	cities := schema.NewTable("cities")
	cities.AddField(nameField)
	cities.AddField(countryField)
	cities.AddField(populationField)

	db.AddTable(cities)

	results := []sql.Term{}
	results = append(results, &sql.Star{})

	sources := []sql.Term{}
	sources = append(sources, &sql.TableReference{Table: cities})

	logicalAnd := &sql.Comparison{
		A:         &sql.Comparison{A: sql.NewFieldReference(countryField, cities, nil), B: sql.NewConstant(schema.SQLVarChar, "NZ"), Operation: sql.OperationEQ},
		B:         &sql.Comparison{A: sql.NewFieldReference(populationField, cities, nil), B: sql.NewConstant(schema.SQLInt, "500000"), Operation: sql.OperationLT},
		Operation: sql.OperationAND,
	}

	test := sql.Select{
		Results: results,
		Sources: sources,
		Where:   logicalAnd,
	}

	fmt.Printf(test.ToSQL(false) + "\n")
}
