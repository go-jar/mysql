package mysql

import (
	"database/sql"
	"errors"
	"github.com/go-jar/golog"
	"reflect"
)

type SimpleOrm struct {
	pool *Pool
	dao  *Dao

	idGenerator *IdGenerator
	traceId     []byte
	logger      golog.ILogger
	useIdGen    bool
}

func NewSimpleOrm(traceId []byte, pool *Pool, useIdGen bool) *SimpleOrm {
	return &SimpleOrm{
		pool:     pool,
		traceId:  traceId,
		logger:   new(golog.NoopLogger),
		useIdGen: useIdGen,
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

func (so *SimpleOrm) Dao() *Dao {
	if so.dao == nil {
		so.dao = &Dao{}
	}

	if so.dao.Client == nil {
		so.dao.Client, _ = so.pool.Get()
		so.dao.Client.SetLogger(so.logger).SetTraceId(so.traceId)
	}

	return so.dao
}

func (so *SimpleOrm) IdGenerator() *IdGenerator {
	if !so.useIdGen {
		return nil
	}

	if so.idGenerator == nil {
		so.idGenerator = NewIdGenerator(so.Dao().Client)
	}

	return so.idGenerator
}

func (so *SimpleOrm) SetTraceId(traceId []byte) *SimpleOrm {
	so.traceId = traceId
	return so
}

func (so *SimpleOrm) SetLogger(logger golog.ILogger) *SimpleOrm {
	so.logger = logger
	return so
}

func (so *SimpleOrm) Renew(traceId []byte, pool *Pool) *SimpleOrm {
	if so.dao != nil && so.dao.Client != nil {
		so.PutBackClient()
	}

	so.traceId = traceId
	so.pool = pool

	return so
}

func (so *SimpleOrm) SetPool(pool *Pool) *SimpleOrm {
	return so.Renew(so.traceId, pool)
}

func (so *SimpleOrm) PutBackClient() {
	if !so.dao.Client.IsClosed() {
		so.dao.Client.SetLogger(new(golog.NoopLogger))
		_ = so.pool.Put(so.dao.Client)
	}

	so.dao.Client = nil
	if so.idGenerator != nil {
		so.idGenerator.SetClient(nil)
	}
}

func (so *SimpleOrm) FillEntityForInsert(rev reflect.Value, entityName, idFieldName string) (int64, error) {
	var id int64
	var err error

	if so.useIdGen {
		id, err = so.IdGenerator().GenerateId(entityName)
		if err != nil {
			return -1, err
		}
		rev.FieldByName(idFieldName).SetInt(id)
	}

	return id, nil
}

func (so *SimpleOrm) Insert(tableName, entityName, idFieldName string, entities ...interface{}) ([]int64, error) {
	cnt := len(entities)
	if cnt <= 0 {
		return nil, errors.New("no values to be inserted")
	}

	colsValues := make([][]interface{}, cnt)
	var ids []int64

	for i, entity := range entities {
		rev := reflect.ValueOf(entity)
		if rev.Kind() == reflect.Ptr {
			rev = rev.Elem()
		}

		id, err := so.FillEntityForInsert(rev, entityName, idFieldName)
		if err != nil {
			return nil, err
		}

		ids = append(ids, id)
		colsValues[i] = ReflectInsertColValues(rev)
	}

	entity := entities[0]
	ret := reflect.TypeOf(entity)
	colNames := ReflectColNames(ret)

	execResult := so.Dao().Insert(tableName, colNames, colsValues...)

	defer so.PutBackClient()

	if execResult.Err != nil {
		return nil, execResult.Err
	}

	return ids, nil
}

func (so *SimpleOrm) GetById(tableName string, id int64, entityPtr interface{}) (bool, error) {
	scanValues := ReflectEntityScanValues(reflect.ValueOf(entityPtr).Elem())

	err := so.Dao().SelectById(tableName, "*", id).Scan(scanValues...)
	defer so.PutBackClient()

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (so *SimpleOrm) UpdateById(tableName string, id int64, newEntityPtr interface{}, updateFields map[string]bool) ([]*QueryItem, error) {
	rev := reflect.ValueOf(newEntityPtr).Elem()
	oldEntity := reflect.New(rev.Type()).Interface()

	find, err := so.GetById(tableName, id, oldEntity)
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

	result := so.Dao().UpdateById(tableName, id, setItems...)
	defer so.PutBackClient()

	if result.Err != nil {
		return nil, result.Err
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}

	return setItems, nil
}

func (so *SimpleOrm) ListByIds(tableName string, ids []int64, orderBy string, entityType reflect.Type, listPtr interface{}) error {
	rows, err := so.Dao().SelectByIds(tableName, "*", orderBy, ids...)
	defer so.PutBackClient()

	if err != nil {
		return err
	}

	return ReflectQueryRowsToEntityList(rows, entityType, listPtr)
}

func (so *SimpleOrm) ListByIdsLimit(tableName string, ids []int64, orderBy string, offset, limit int64, entityType reflect.Type, listPtr interface{}) error {
	rows, err := so.Dao().SelectByIdsLimit(tableName, "*", orderBy, offset, limit, ids...)
	defer so.PutBackClient()

	if err != nil {
		return err
	}

	return ReflectQueryRowsToEntityList(rows, entityType, listPtr)
}

func (so *SimpleOrm) SimpleQueryAnd(tableName string, qp *QueryParams, entityType reflect.Type, listPtr interface{}) error {
	var setItems []*QueryItem

	if qp != nil && qp.ParamsStructPtr != nil {
		setItems = ReflectQueryItems(reflect.ValueOf(qp.ParamsStructPtr).Elem(), qp.Required, qp.Conditions)
	}

	rows, err := so.Dao().SimpleSelectAnd(tableName, "*", qp.OrderBy, qp.Offset, qp.Cnt, setItems...)

	defer so.PutBackClient()

	if err != nil {
		return err
	}

	return ReflectQueryRowsToEntityList(rows, entityType, listPtr)
}

func (so *SimpleOrm) SimpleTotalAnd(tableName string, qp *QueryParams) (int64, error) {
	var items []*QueryItem
	if qp != nil && qp.ParamsStructPtr != nil {
		items = ReflectQueryItems(reflect.ValueOf(qp.ParamsStructPtr).Elem(), qp.Required, qp.Conditions)
	}

	total, err := so.Dao().SelectTotalAnd(tableName, items...)
	defer so.PutBackClient()

	return total, err
}
