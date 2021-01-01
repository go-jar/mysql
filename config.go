package mysql

import (
	"time"

	"github.com/go-jar/golog"
	"github.com/go-sql-driver/mysql"
)

const (
	DefaultConnectTimeout = 10 * time.Second
	DefaultReadTimeout    = 10 * time.Second
	DefaultWriteTimeout   = 10 * time.Second
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
		Timeout:              DefaultConnectTimeout,
		ReadTimeout:          DefaultReadTimeout,
		WriteTimeout:         DefaultWriteTimeout,
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	return &Config{
		Config: config,

		LogLevel: golog.LevelInfo,
	}
}
