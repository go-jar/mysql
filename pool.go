package mysql

import (
	"github.com/go-jar/pool"
)

type Pool struct {
	pl            *pool.Pool
	NewClientFunc func() (*Client, error)
}

type NewClientFunc func() (*Client, error)

func NewPool(config *pool.Config, ncf NewClientFunc) *Pool {
	p := &Pool{
		pl:            pool.NewPool(config, nil),
		NewClientFunc: ncf,
	}
	p.pl.NewItemFunc = p.newConn
	return p
}

func (p *Pool) Get() (*Client, error) {
	conn, err := p.pl.Get()
	if err != nil {
		return nil, err
	}
	return conn.(*Client), err
}

func (p *Pool) Put(client *Client) error {
	return p.pl.Put(client)
}

func (p *Pool) newConn() (pool.IConn, error) {
	return p.NewClientFunc()
}
