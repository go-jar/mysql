package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/goinbox/gomisc"
)

type Orm struct {
	Dao *Dao
}

func NewOrm(client *Client) *Orm {
	return &Orm{
		Dao: NewDao(client),
	}
}

type QueryParams struct {
	ParamsStructPtr interface{}
	Required        map[string]bool
	Conditions      map[string]string

	OrderBy string
	Offset  int64
	Cnt     int64
}

func (o *Orm) Insert(tableName string, entities ...interface{}) error {
	cnt := len(entities)
	if cnt <= 0 {
		return errors.New("no values to be inserted")
	}

	colsValues := make([][]interface{}, cnt)

	for i, entity := range entities {
		rev := reflect.ValueOf(entity)
		if rev.Kind() == reflect.Ptr {
			rev = rev.Elem()
		}

		colsValues[i] = ReflectInsertColValues(rev)
	}

	entity := entities[0]
	ret := reflect.TypeOf(entity)
	fmt.Println(ret)
	colNames := ReflectColNames(ret)

	err := o.Dao.Insert(tableName, colNames, colsValues...).Err

	if err != nil {
		return err
	}

	return nil
}

func (o *Orm) GetById(tableName string, id int64, entityPtr interface{}) (bool, error) {
	scanValues := ReflectEntityScanValues(reflect.ValueOf(entityPtr).Elem())

	err := o.Dao.SelectById(tableName, "*", id).Scan(scanValues...)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (o *Orm) UpdateById(tableName string, id int64, newEntityPtr interface{}, updateFields map[string]bool) ([]*QueryItem, error) {
	rev := reflect.ValueOf(newEntityPtr).Elem()
	oldEntity := reflect.New(rev.Type()).Interface()

	find, err := o.GetById(tableName, id, oldEntity)
	if err != nil {
		return nil, err
	}
	if !find {
		return nil, nil
	}

	setItems := ReflectUpdateItems(reflect.ValueOf(oldEntity).Elem(), rev, updateFields)
	if len(setItems) == 0 {
		return nil, nil
	}

	setItems = append(setItems, NewQueryItem("edit_time", "", time.Now().Format(gomisc.TimeGeneralLayout())))
	result := o.Dao.UpdateById(tableName, id, setItems...)

	if result.Err != nil {
		return nil, result.Err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return setItems, nil
}

func (o *Orm) ListByIds(tableName string, ids []int64, orderBy string, entityType reflect.Type, listPtr interface{}) error {
	rows, err := o.Dao.SelectByIds(tableName, "*", orderBy, ids...)

	if err != nil {
		return err
	}

	return ReflectQueryRowsToEntityList(rows, entityType, listPtr)
}

func (o *Orm) SimpleQueryAnd(tableName string, qp *QueryParams, entityType reflect.Type, listPtr interface{}) error {
	var setItems []*QueryItem
	if qp != nil && qp.ParamsStructPtr != nil {
		setItems = ReflectQueryItems(reflect.ValueOf(qp.ParamsStructPtr).Elem(), qp.Required, qp.Conditions)
	}

	rows, err := o.Dao.SimpleSelectAnd(tableName, "*", qp.OrderBy, qp.Offset, qp.Cnt, setItems...)

	if err != nil {
		return err
	}

	return ReflectQueryRowsToEntityList(rows, entityType, listPtr)
}

func (o *Orm) SimpleTotalAnd(tableName string, qp *QueryParams) (int64, error) {
	var items []*QueryItem
	if qp != nil && qp.ParamsStructPtr != nil {
		items = ReflectQueryItems(reflect.ValueOf(qp.ParamsStructPtr).Elem(), qp.Required, qp.Conditions)
	}

	total, err := o.Dao.SelectTotalAnd(tableName, items...)

	return total, err
}
