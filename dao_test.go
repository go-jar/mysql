package mysql

import (
	"fmt"
	"testing"
)

func TestDaoWrite(t *testing.T) {
	dao := NewDao(client)

	colNames := []string{"id", "name", "age"}
	colsValues := [][]interface{}{
		{10, "d", 10},
		{11, "e", 11},
		{12, "f", 12},
		{13, "g", 13},
		{14, "h", 14},
		{15, "i", 15},
	}

	result := dao.Insert(TABLE_NAME, colNames, colsValues...)
	printResult("Insert: ", result)

	setItems := []*QueryItem{
		NewPair("name", "dd"),
		NewPair("age", 11),
	}
	result = dao.UpdateById(TABLE_NAME, 10, setItems...)
	printResult("UpdateById: ", result)

	setItems = []*QueryItem{
		NewPair("name", "ee"),
		NewPair("age", 11),
	}
	ids := []int64{11, 12}
	result = dao.UpdateByIds(TABLE_NAME, ids, setItems...)
	printResult("UpdateByIds: ", result)

	result = dao.DeleteById(TABLE_NAME, 13)
	printResult("DeleteById: ", result)

	result = dao.DeleteByIds(TABLE_NAME, 14, 15)
	printResult("DeleteByIds: ", result)
}

func TestDaoRead(t *testing.T) {
	dao := NewDao(client)
	item := new(DemoItem)

	row := dao.SelectById(TABLE_NAME, "*", 11)
	row.Scan(&item.Id, &item.Name, &item.Name)
	printResult("SelectById: ", item)

	fmt.Println("SelectByIds: ")
	rows, _ := dao.SelectByIds(TABLE_NAME, "*", "age", 11, 12)
	for rows.Next() {
		rows.Scan(&item.Id, &item.Name, &item.Name)
		fmt.Println(item)
	}
	fmt.Println("====================")

	conditions := []*QueryItem{
		NewCondition("name", EQUAL, "ee"),
		NewCondition("age", EQUAL, 11),
	}

	fmt.Println("SimpleSelectAnd: ")
	rows, _ = dao.SimpleSelectAnd(TABLE_NAME, "*", "id desc", 0, 10, conditions...)
	for rows.Next() {
		rows.Scan(&item.Id, &item.Name, &item.Age)
		fmt.Println(item)
	}
	fmt.Println("====================")

	total, _ := dao.SelectTotalAnd(TABLE_NAME, conditions...)
	printResult("SelectTotalAnd: ", total)

	conditions = []*QueryItem{
		NewCondition("name", EQUAL, "dd"),
		NewCondition("name", EQUAL, "ee"),
	}

	fmt.Println("SimpleSelectOr: ")
	rows, _ = dao.SimpleSelectOr(TABLE_NAME, "*", "id desc", 0, 10, conditions...)
	for rows.Next() {
		rows.Scan(&item.Id, &item.Name, &item.Age)
		fmt.Println(item)
	}
	fmt.Println("====================")

	total, _ = dao.SelectTotalOr(TABLE_NAME, conditions...)
	printResult("SelectTotalOr: ", total)
}

func printResult(msg string, result interface{}) {
	fmt.Println(msg)
	fmt.Println(result)
	fmt.Println("====================")
}
