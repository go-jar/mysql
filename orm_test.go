package mysql

import (
	"fmt"
	"github.com/go-jar/operator"
	"reflect"
	"testing"
)

/*
CREATE TABLE `demo` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(20) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `status`varchar(20) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `add_time` datetime,
  `edit_time` datetime,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=34 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
*/

type SqlBaseEntity struct {
	Id       int64  `mysql:"id" json:"id"`
	AddTime  string `mysql:"add_time" json:"add_time"`
	EditTime string `mysql:"edit_time" json:"edit_time"`
}

type demoEntity struct {
	SqlBaseEntity

	Name   string `mysql:"name" json:"name"`
	Status int    `mysql:"status" json:"status"`
}

func TestSqlSvcInsertGetListUpdateDelete(t *testing.T) {
	client := getTestClient()
	ss := NewOrm(client)

	item := &demoEntity{
		Name:   "tdj",
		Status: 1,
	}

	fmt.Println("test Insert")

	tableName := "demo"
	err := ss.Insert(tableName, item)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("test List")

	ids := []int64{0, 1, 2}
	var data []*demoEntity
	demoEntityType := reflect.TypeOf(demoEntity{})
	err = ss.ListByIds(tableName, ids, "add_time desc", demoEntityType, &data)
	if err != nil {
		fmt.Println(err)
	}
	for i, item := range data {
		fmt.Println(i, item)
	}

	sqp := &QueryParams{
		ParamsStructPtr: &demoEntity{
			Status: 1,
		},
		Required:   map[string]bool{"status": true},
		Conditions: map[string]string{"status": operator.EQUAL},

		OrderBy: "add_time desc",
		Offset:  0,
		Cnt:     10,
	}
	cnt, err := ss.SimpleTotalAnd("demo", sqp)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(cnt)

	data = []*demoEntity{}
	err = ss.SimpleQueryAnd(tableName, sqp, demoEntityType, &data)
	if err != nil {
		fmt.Println(err)
	}
	for i, item := range data {
		fmt.Println(i, item)
	}

	newDemo := &demoEntity{
		Name: "new-demo",
	}
	updateFields := map[string]bool{"name": true}
	setItems, err := ss.UpdateById(tableName, ids[0], newDemo, updateFields)
	if err != nil {
		fmt.Println(err)
	}
	for i, item := range setItems {
		fmt.Println(i, item)
	}

	fmt.Println("test Get")

	item = &demoEntity{}
	find, err := ss.GetById(tableName, ids[0], item)
	if !find {
		fmt.Println("not found")
	}
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(item)

	fmt.Println("test Delete")

	result := ss.Dao.DeleteById(tableName, ids[0])

	fmt.Println(result)
}
