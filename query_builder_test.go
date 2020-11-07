package mysql

import (
	"fmt"
	"github.com/go-jar/operator"
	"testing"
)

const TABLE_NAME = "people"

var qb QueryBuilder

func TestInsert(t *testing.T) {
	qb.Insert(TABLE_NAME, "name", "age").
		Values(
			[]interface{}{"c", 5},
			[]interface{}{"d", 6})

	printQueryAndArgs()
}

func TestSelect(t *testing.T) {
	qb.Select(TABLE_NAME, "*, count(*)").
		WhereAnd(
			NewCondition("name", operator.EQUAL, "c"),
			NewCondition("name", operator.LIKE, "c%")).
		GroupBy("name").
		HavingAnd(
			NewCondition("age", operator.GREATER_EQUAL, 0),
			NewCondition("age", operator.LESS_EQUAL, 10)).
		OrderBy("age").
		Limit(0, 10)

	printQueryAndArgs()
}

func TestUpdate(t *testing.T) {
	qb.Update(TABLE_NAME).
		Set(NewPair("name", "e"), NewPair("age", 7)).
		WhereOr(
			NewCondition("name", operator.EQUAL, "a"),
			NewCondition("age", operator.LESS_EQUAL, 6))

	printQueryAndArgs()
}

func TestDelete(t *testing.T) {
	qb.Delete(TABLE_NAME).
		WhereAnd(
			NewCondition("name", operator.IN, []string{"cc", "dd"}))

	printQueryAndArgs()
}

func printQueryAndArgs() {
	fmt.Println(qb.Query(), qb.Args())
}
