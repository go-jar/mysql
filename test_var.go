package mysql

import "github.com/go-jar/golog"

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

type DemoItem struct {
	Id   int64
	Name string
	Age  string
}

func getTestClient() *Client {
	config := NewConfig("root", "passwd", "127.0.0.1", "3306", "demo")

	logger, _ := golog.NewConsoleLogger(golog.LEVEL_INFO)
	client, _ := NewClient(config, logger)

	return client
}
