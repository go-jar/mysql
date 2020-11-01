package mysql

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/go-jar/golog"
)

/*
create database demo;

use demo;

CREATE TABLE `people` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(20) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `age`  tinyint(4) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
*/

var client *Client

type DemoItem struct {
	Id   int64
	Name string
	Age  string
}

func init() {
	logger, _ := golog.NewConsoleLogger(golog.LEVEL_INFO)
	config := NewConfig("root", "yuntest#cloud", "10.66.172.152", "3306", "demo")
	client, _ = NewClient(config, logger)
}

func TestClient_Exec(t *testing.T) {
	result, err := client.Exec("insert into people(name, age) values (?, ?), (?, ?)", "a", 1, "b", 2)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		lastInsertId, err := result.LastInsertId()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("LastInsertId: " + strconv.FormatInt(lastInsertId, 10))
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("RawsAffected: " + strconv.FormatInt(rowsAffected, 10))
		}
	}
}

func TestClient_Query(t *testing.T) {
	raws, err := client.Query("select * from people where name in (?, ?)", "a", "b")

	if err != nil {
		fmt.Println(err.Error())
	} else {
		for raws.Next() {
			item := new(DemoItem)
			err = raws.Scan(&item.Id, &item.Name, &item.Age)
			if err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println(item)
			}
		}
	}
}

func TestClient_QueryRow(t *testing.T) {
	raw := client.QueryRow("select * from people where name = ?", "a")
	item := new(DemoItem)
	err := raw.Scan(&item.Id, &item.Name, &item.Age)

	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(item)
	}
}

func TestClient_Trans(t *testing.T) {
	client.Begin()
	client.Exec("update people set ? = ? where name = ?", "age", 3, "a")
	client.Exec("update people set age = ? where name = ?", 4, "b")
	client.Commit()

	client.Begin()
	client.Exec("update demo people age = ? where name = ?", 1, "a")
	err := client.Rollback()
	if err != nil {
		fmt.Println(err.Error())
	}
}
