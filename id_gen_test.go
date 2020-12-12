package mysql

import (
	"fmt"
	"testing"
)

/*
CREATE TABLE IF NOT EXISTS `id_gen` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(50) COLLATE utf8mb4_bin NOT NULL DEFAULT '',
  `max_id` bigint(20) unsigned NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`),
  UNIQUE KEY `name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

INSERT INTO id_gen (name, max_id) VALUES ('demo', 0);
*/

func TestSqlIdGen(t *testing.T) {
	client, _ := newMysqlTestClient()
	idGen := NewIdGenerator(client)

	for i := 0; i < 10; i++ {
		id, err := idGen.GenerateId("demo")
		fmt.Println(id, err)
	}
}
