package mysql

import (
	"fmt"
	"github.com/go-jar/golog"
	"reflect"
	"testing"
	"time"
)

/*
CREATE TABLE `demo` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `add_time` datetime,
  `edit_time` datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp(),
  `name` varchar(20) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status`varchar(20) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=0 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
*/

type SqlBaseEntity struct {
	Id       int64     `mysql:"id" json:"id"`
	AddTime  time.Time `mysql:"add_time" json:"add_time"`
	EditTime time.Time `mysql:"edit_time" json:"edit_time"`
}

type demoEntity struct {
	SqlBaseEntity

	Name   string `mysql:"name" json:"name"`
	Status int    `mysql:"status" json:"status"`
}

func TestOrmInsertGetListUpdateDelete(t *testing.T) {
	config := &PoolConfig{NewClientFunc: newMysqlTestClient}
	config.MaxConns = 100
	config.MaxIdleTime = time.Second * 5

	pool := NewPool(config)
	logger, _ := golog.NewConsoleLogger(golog.LevelInfo)
	orm := NewSimpleOrm([]byte("-"), pool, true).SetLogger(logger)

	item := &demoEntity{
		Name:   "tdj",
		Status: 1,
		SqlBaseEntity: SqlBaseEntity{
			AddTime:  time.Now(),
			EditTime: time.Now(),
		},
	}

	fmt.Println("========test Insert")
	tableName := "demo"
	ids, err := orm.Insert(tableName, tableName, "Id", item)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(ids)
	}

	fmt.Println("========test List")

	var data []*demoEntity
	demoEntityType := reflect.TypeOf(demoEntity{})
	err = orm.ListByIds(tableName, ids, "id desc", demoEntityType, &data)
	if err != nil {
		fmt.Println(err)
	} else {
		for i, item := range data {
			fmt.Println(i, item)
		}
	}

	fmt.Println("========test SimpleTotalAnd")

	qp := &QueryParams{
		ParamsStructPtr: &demoEntity{
			Status: 1,
		},
		Required:   map[string]bool{"status": true},
		Conditions: map[string]string{"status": CondEqual},

		OrderBy: "id desc",
		Offset:  0,
		Cnt:     10,
	}
	cnt, err := orm.SimpleTotalAnd("demo", qp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cnt)

	fmt.Println("========test SimpleQueryAnd")

	data = []*demoEntity{}
	err = orm.SimpleQueryAnd(tableName, qp, demoEntityType, &data)
	if err != nil {
		fmt.Println(err)
	}
	for i, item := range data {
		fmt.Println(i, item)
	}

	fmt.Println("========test UpdateById")

	newDemo := &demoEntity{
		Name: "new-demo",
	}
	updateFields := map[string]bool{"name": true}
	setItems, err := orm.UpdateById(tableName, ids[0], newDemo, updateFields)
	if err != nil {
		fmt.Println(err)
	}
	for i, item := range setItems {
		fmt.Println(i, item)
	}

	fmt.Println("========test Get")

	item = &demoEntity{}
	find, err := orm.GetById(tableName, ids[0], item)
	if !find {
		fmt.Println("not found")
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(item)

	fmt.Println("========test Delete")

	result := orm.Dao().DeleteById(tableName, ids[0])

	fmt.Println(result)
}
