package mysql

import (
	"database/sql"
	"errors"
	"reflect"

	"github.com/go-jar/golog"
)

type Orm struct {
	pool *Pool
	dao  *Dao

	idGenerator *IdGenerator
	traceId     []byte
	logger      golog.ILogger
}

func NewOrm(pool *Pool) *Orm {
	return &Orm{
		pool:   pool,
		logger: new(golog.NoopLogger),
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

func (o *Orm) Dao() *Dao {
	if o.dao == nil {
		o.dao = &Dao{}
	}

	if o.dao.Client == nil {
		o.dao.Client, _ = o.pool.Get()
		o.dao.Client.SetLogger(o.logger).SetTraceId(o.traceId)
	}

	return o.dao
}

func (o *Orm) IdGenerator() *IdGenerator {
	if o.idGenerator == nil {
		o.idGenerator = NewIdGenerator(o.Dao().Client)
	}

	return o.idGenerator
}

func (o *Orm) SetTraceId(traceId []byte) *Orm {
	o.traceId = traceId
	return o
}

func (o *Orm) SetLogger(logger golog.ILogger) *Orm {
	o.logger = logger
	return o
}

func (o *Orm) Renew(traceId []byte, pool *Pool) *Orm {
	if o.dao != nil && o.dao.Client != nil {
		o.PutBackClient()
	}

	o.traceId = traceId
	o.pool = pool

	return o
}

func (o *Orm) SetPool(pool *Pool) *Orm {
	return o.Renew(o.traceId, pool)
}

func (o *Orm) PutBackClient() {
	if !o.dao.Client.IsClosed() {
		o.dao.Client.SetLogger(new(golog.NoopLogger))
		_ = o.pool.Put(o.dao.Client)
	}

	o.dao.Client = nil
	if o.idGenerator != nil {
		o.idGenerator.SetClient(nil)
	}
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
	colNames := ReflectColNames(ret)
	err := o.Dao().Insert(tableName, colNames, colsValues...).Err
	defer o.PutBackClient()

	if err != nil {
		return err
	}

	return nil
}

func (o *Orm) GetById(tableName string, id int64, entityPtr interface{}) (bool, error) {
	scanValues := ReflectEntityScanValues(reflect.ValueOf(entityPtr).Elem())

	err := o.Dao().SelectById(tableName, "*", id).Scan(scanValues...)
	defer o.PutBackClient()

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

	result := o.Dao().UpdateById(tableName, id, setItems...)
	defer o.PutBackClient()

	if result.Err != nil {
		return nil, result.Err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return setItems, nil
}

func (o *Orm) ListByIds(tableName string, ids []int64, orderBy string, entityType reflect.Type, listPtr interface{}) error {
	rows, err := o.Dao().SelectByIds(tableName, "*", orderBy, ids...)
	defer o.PutBackClient()

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

	rows, err := o.Dao().SimpleSelectAnd(tableName, "*", qp.OrderBy, qp.Offset, qp.Cnt, setItems...)

	defer o.PutBackClient()

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

	total, err := o.Dao().SelectTotalAnd(tableName, items...)
	defer o.PutBackClient()

	return total, err
}
