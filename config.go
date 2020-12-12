package mysql

import (
	"time"

	"github.com/go-jar/golog"
	"github.com/go-sql-driver/mysql"
)

const (
	DEFAULT_CONNECT_TIMEOUT = 10 * time.Second
	DEFAULT_READ_TIMEOUT    = 10 * time.Second
	DEFAULT_WRITE_TIMEOUT   = 10 * time.Second
)

type Config struct {
	*mysql.Config

	LogLevel int
}

func NewConfig(user, passwd, host, port, dbName string) *Config {
	params := map[string]string{
		"interpolateParams": "true",
	}

	config := &mysql.Config{
		User:                 user,
		Passwd:               passwd,
		Net:                  "tcp",
		Addr:                 host + ":" + port,
		DBName:               dbName,
		Params:               params,
		Timeout:              DEFAULT_CONNECT_TIMEOUT,
		ReadTimeout:          DEFAULT_READ_TIMEOUT,
		WriteTimeout:         DEFAULT_WRITE_TIMEOUT,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	return &Config{
		Config: config,

		LogLevel: golog.LEVEL_INFO,
	}
}
