package mysql

import (
	"reflect"
	"strings"
)

const (
	CondEqual        = "="
	CondNotEqual     = "!="
	CondLess         = "<"
	CondLessEqual    = "<="
	CondGreater      = ">"
	CondGreaterEqual = ">="
	CondIn           = "in"
	CondNotIn        = "not in"
	CondLike         = "like"
	CondBetween      = "between"
)

type QueryItem struct {
	Name      string
	Condition string
	Value     interface{}
}

func NewCondition(name, condition string, value interface{}) *QueryItem {
	return &QueryItem{
		Name:      name,
		Condition: condition,
		Value:     value,
	}
}

func NewQueryItem(name, condition string, value interface{}) *QueryItem {
	return &QueryItem{
		Name:      name,
		Condition: condition,
		Value:     value,
	}
}

func NewPair(name string, value interface{}) *QueryItem {
	return &QueryItem{
		Name:  name,
		Value: value,
	}
}

type QueryBuilder struct {
	query string
	args  []interface{}
}

func (qb *QueryBuilder) Query() string {
	return qb.query
}

func (qb *QueryBuilder) Args() []interface{} {
	return qb.args
}

func (qb *QueryBuilder) Insert(tableName string, columnNames ...string) *QueryBuilder {
	qb.args = nil
	qb.query = "insert into " + tableName + " ("
	qb.query += strings.Join(columnNames, ", ") + ") values "
	return qb
}

func (qb *QueryBuilder) Values(values ...[]interface{}) *QueryBuilder {
	rawNum := len(values) - 1
	if rawNum == -1 {
		return qb
	}

	for i := 0; i < rawNum; i++ {
		qb.buildInsertRow(values[i])
		qb.query += ", "
	}

	qb.buildInsertRow(values[rawNum])
	return qb
}

func (qb *QueryBuilder) Delete(tableName string) *QueryBuilder {
	qb.args = nil
	qb.query = "delete from " + tableName
	return qb
}

func (qb *QueryBuilder) Update(tableName string) *QueryBuilder {
	qb.args = nil
	qb.query = "update " + tableName
	return qb
}

func (qb *QueryBuilder) Set(items ...*QueryItem) *QueryBuilder {
	n := len(items) - 1
	if n == -1 {
		return qb
	}

	qb.query += " set "

	for i := 0; i < n; i++ {
		qb.query += items[i].Name + " = ?, "
		qb.args = append(qb.args, items[i].Value)
	}
	qb.query += items[n].Name + " = ? "
	qb.args = append(qb.args, items[n].Value)

	return qb
}

func (qb *QueryBuilder) Select(tableName, what string) *QueryBuilder {
	qb.args = nil
	qb.query = "select " + what + " from " + tableName
	return qb
}

func (qb *QueryBuilder) WhereAnd(conditions ...*QueryItem) *QueryBuilder {
	if len(conditions) == 0 {
		return qb
	}

	qb.query += " where "
	qb.buildCondition("and", conditions...)
	return qb
}

func (qb *QueryBuilder) WhereOr(conditions ...*QueryItem) *QueryBuilder {
	if len(conditions) == 0 {
		return qb
	}

	qb.query += " where "
	qb.buildCondition("or", conditions...)
	return qb
}

func (qb *QueryBuilder) OrderBy(orderBy string) *QueryBuilder {
	if orderBy != "" {
		qb.query += " order by " + orderBy
	}
	return qb
}

func (qb *QueryBuilder) GroupBy(columnNames string) *QueryBuilder {
	if columnNames != "" {
		qb.query += " group by " + columnNames
	}
	return qb
}

func (qb *QueryBuilder) HavingAnd(conditions ...*QueryItem) *QueryBuilder {
	if len(conditions) == 0 {
		return qb
	}

	qb.query += " having "
	qb.buildCondition("and", conditions...)
	return qb
}

func (qb *QueryBuilder) HavingOr(conditions ...*QueryItem) *QueryBuilder {
	if len(conditions) == 0 {
		return qb
	}

	qb.query += " having "
	qb.buildCondition("or", conditions...)
	return qb
}

func (qb *QueryBuilder) Limit(offset, cnt int64) *QueryBuilder {
	if offset < 0 || cnt <= 0 {
		return qb
	}

	qb.query += " limit ?, ?"
	qb.args = append(qb.args, offset, cnt)

	return qb
}

func (qb *QueryBuilder) buildInsertRow(args []interface{}) {
	colNum := len(args) - 1
	if colNum == -1 {
		return
	}

	qb.query += "("

	for i := 0; i < colNum; i++ {
		qb.query += "?, "
		qb.args = append(qb.args, args[i])
	}

	qb.query += "?)"
	qb.args = append(qb.args, args[colNum])
}

func (qb *QueryBuilder) buildCondition(andOr string, conditions ...*QueryItem) {
	n := len(conditions) - 1
	if n == -1 {
		return
	}

	for i := 0; i < n; i++ {
		qb.buildConditionWhere(conditions[i])
		qb.query += " " + andOr + " "
	}
	qb.buildConditionWhere(conditions[n])
}

func (qb *QueryBuilder) buildConditionWhere(condition *QueryItem) {
	switch condition.Condition {
	case CondEqual, CondNotEqual, CondLess, CondLessEqual, CondGreater, CondGreaterEqual:
		qb.query += condition.Name + " " + condition.Condition + " ? "
		qb.args = append(qb.args, condition.Value)
	case CondLike:
		qb.query += condition.Name + " like ?"
		qb.args = append(qb.args, condition.Value)
	case CondBetween:
		qb.query += condition.Name + " between ? and ?"
		rev := reflect.ValueOf(condition.Value)
		qb.args = append(qb.args, rev.Index(0).Interface(), rev.Index(1).Interface())
	case CondIn:
		qb.buildConditionInOrNot("in", condition)
	case CondNotIn:
		qb.buildConditionInOrNot("not in", condition)
	}
}

func (qb *QueryBuilder) buildConditionInOrNot(inOrNot string, condition *QueryItem) {
	rev := reflect.ValueOf(condition.Value)
	n := rev.Len() - 1
	if n == -1 {
		return
	}

	qb.query += condition.Name + " " + inOrNot + " ("

	for i := 0; i < n; i++ {
		qb.query += "?, "
		qb.args = append(qb.args, rev.Index(i).Interface())
	}
	qb.query += "?)"
	qb.args = append(qb.args, rev.Index(n).Interface())
}
