package mysql

import (
	"database/sql"
)

type ExecResult struct {
	Err          error
	LastInsertId int64
	RowsAffected int64
}

type Dao struct {
	*Client
}

func NewDao(client *Client) *Dao {
	return &Dao{
		Client: client,
	}
}

func (d *Dao) Insert(tableName string, colNames []string, colValues ...[]interface{}) *ExecResult {
	qb := new(QueryBuilder)
	qb.Insert(tableName, colNames...).
		Values(colValues...)

	return GetExecResult(d.Exec(qb.Query(), qb.Args()...))
}

func (d *Dao) DeleteById(tableName string, id int64) *ExecResult {
	qb := new(QueryBuilder)
	qb.Delete(tableName).
		WhereAnd(NewCondition("id", CondEqual, id))

	return GetExecResult(d.Exec(qb.Query(), qb.Args()...))
}

func (d *Dao) DeleteByIds(tableName string, ids ...int64) *ExecResult {
	qb := new(QueryBuilder)
	qb.Delete(tableName).
		WhereAnd(NewCondition("id", CondIn, ids))

	return GetExecResult(d.Exec(qb.Query(), qb.Args()...))
}

func (d *Dao) UpdateById(tableName string, id int64, pairs ...*QueryItem) *ExecResult {
	qb := new(QueryBuilder)
	qb.Update(tableName).
		Set(pairs...).
		WhereAnd(NewCondition("id", CondEqual, id))

	return GetExecResult(d.Exec(qb.Query(), qb.Args()...))
}

func (d *Dao) UpdateByIds(tableName string, ids []int64, pairs ...*QueryItem) *ExecResult {
	qb := new(QueryBuilder)
	qb.Update(tableName).
		Set(pairs...).
		WhereAnd(NewCondition("id", CondIn, ids))

	return GetExecResult(d.Exec(qb.Query(), qb.Args()...))
}

func (d *Dao) SelectById(tableName, what string, id int64) *sql.Row {
	qb := new(QueryBuilder)
	qb.Select(tableName, what).
		WhereAnd(NewCondition("id", CondEqual, id))

	return d.QueryRow(qb.Query(), qb.Args()...)
}

func (d *Dao) SelectByIds(tableName, what, orderBy string, ids ...int64) (*sql.Rows, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, what).
		WhereAnd(NewCondition("id", CondIn, ids)).
		OrderBy(orderBy)

	return d.Query(qb.Query(), qb.Args()...)
}

func (d *Dao) SelectByIdsLimit(tableName, what, orderBy string, offset, limit int64, ids ...int64) (*sql.Rows, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, what).
		WhereAnd(NewCondition("id", CondIn, ids)).
		OrderBy(orderBy).
		Limit(offset, limit)

	return d.Query(qb.Query(), qb.Args()...)
}

func (d *Dao) SelectTotalAnd(tableName string, conditions ...*QueryItem) (int64, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, "count(1)").
		WhereAnd(conditions...)

	var total int64
	err := d.QueryRow(qb.Query(), qb.Args()...).Scan(&total)

	return total, err
}

func (d *Dao) SelectTotalOr(tableName string, conditions ...*QueryItem) (int64, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, "count(1)").
		WhereOr(conditions...)

	var total int64
	err := d.QueryRow(qb.Query(), qb.Args()...).Scan(&total)

	return total, err
}

func (d *Dao) SimpleSelectAnd(tableName, what, orderBy string, offset, limit int64, conditions ...*QueryItem) (*sql.Rows, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, what).
		WhereAnd(conditions...).
		OrderBy(orderBy).
		Limit(offset, limit)

	return d.Query(qb.Query(), qb.Args()...)
}

func (d *Dao) SimpleSelectOr(tableName, what, orderBy string, offset, limit int64, conditions ...*QueryItem) (*sql.Rows, error) {
	qb := new(QueryBuilder)
	qb.Select(tableName, what).
		WhereOr(conditions...).
		OrderBy(orderBy).
		Limit(offset, limit)

	return d.Query(qb.Query(), qb.Args()...)
}

func GetExecResult(result sql.Result, err error) *ExecResult {
	execResult := new(ExecResult)

	if err != nil {
		execResult.Err = err
	} else {
		lid, err := result.LastInsertId()
		if err != nil {
			execResult.Err = err
		} else {
			execResult.LastInsertId = lid
			ra, err := result.RowsAffected()
			if err != nil {
				execResult.Err = err
			} else {
				execResult.RowsAffected = ra
			}
		}
	}

	return execResult
}
