package mysql

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/go-jar/golog"
	"github.com/goinbox/gomisc"
)

type Client struct {
	config *Config

	db *sql.DB
	tx *sql.Tx

	isConnClosed bool

	logger    golog.ILogger
	traceId   []byte
	logPrefix []byte
}

func NewClient(config *Config, logger golog.ILogger) (*Client, error) {
	if config.LogLevel == 0 {
		config.LogLevel = golog.LevelInfo
	}

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, err
	}

	if logger == nil {
		logger = new(golog.NoopLogger)
	}

	return &Client{
		config:    config,
		db:        db,
		tx:        nil,
		logger:    logger,
		traceId:   []byte("-"),
		logPrefix: nil,
	}, nil
}

func (c *Client) SetLogger(logger golog.ILogger) *Client {
	if logger == nil {
		logger = new(golog.NoopLogger)
	}

	c.logger = logger
	return c
}

func (c *Client) SetTraceId(traceId []byte) *Client {
	if traceId == nil {
		traceId = []byte("-")
	}

	c.traceId = traceId
	return c
}

func (c *Client) SetLogPrefix(logPrefix []byte) *Client {
	c.logPrefix = logPrefix
	return c
}

func (c *Client) IsClosed() bool {
	return c.isConnClosed
}

func (c *Client) Free() {
	c.db.Close()
	c.tx = nil
	c.isConnClosed = true
}

func (c *Client) Exec(query string, args ...interface{}) (sql.Result, error) {
	c.log(query, args...)

	if c.tx != nil {
		return c.tx.Exec(query, args...)
	} else {
		return c.db.Exec(query, args...)
	}
}

func (c *Client) Query(query string, args ...interface{}) (*sql.Rows, error) {
	c.log(query, args...)

	if c.tx != nil {
		return c.tx.Query(query, args...)
	} else {
		return c.db.Query(query, args...)
	}
}

func (c *Client) QueryRow(query string, args ...interface{}) *sql.Row {
	c.log(query, args...)

	if c.tx != nil {
		return c.tx.QueryRow(query, args...)
	} else {
		return c.db.QueryRow(query, args...)
	}
}

func (c *Client) Begin() error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	c.log("BEGIN")
	c.tx = tx

	return nil
}

func (c *Client) Commit() error {
	defer func() {
		c.tx = nil
	}()

	if c.tx != nil {
		c.log("COMMIT")

		return c.tx.Commit()
	}

	return errors.New("Not in transaction")
}

func (c *Client) Rollback() error {
	defer func() {
		c.tx = nil
	}()

	if c.tx != nil {
		c.log("ROLLBACK")

		return c.tx.Rollback()
	}

	return errors.New("Not in transaction")
}

func (c *Client) log(query string, args ...interface{}) {
	query = strings.Replace(query, "?", "%s", -1)
	vs := make([]interface{}, len(args))

	for i, v := range args {
		s := fmt.Sprint(v)
		switch v.(type) {
		case string:
			vs[i] = "'" + s + "'"
		default:
			vs[i] = s
		}
	}

	query = fmt.Sprintf(query, vs...)

	var msg []byte
	if c.logPrefix != nil {
		msg = gomisc.AppendBytes(c.traceId, []byte("\t"), c.logPrefix, []byte(query))
	} else {
		msg = gomisc.AppendBytes(c.traceId, []byte("\t"), []byte(query))
	}

	c.logger.Log(c.config.LogLevel, msg)
}
