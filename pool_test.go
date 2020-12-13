package mysql

import (
	"fmt"
	"testing"
	"time"

	"github.com/go-jar/pool"
)

func TestPool(t *testing.T) {
	config := &pool.Config{
		MaxConns:    100,
		MaxIdleTime: time.Second * 5,
	}

	pool := NewPool(config, newMysqlTestClient)

	testPool(pool, t)
	testPool(pool, t)
	time.Sleep(time.Second * 7)
	testPool(pool, t)
}

func testPool(p *Pool, t *testing.T) {
	client, _ := p.Get()
	raw := client.QueryRow("select * from people where id = ?", 10)
	item := new(DemoItem)
	err := raw.Scan(&item.Id, &item.Name, &item.Age)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println(item)
	}

	p.Put(client)
}