package storage

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Client struct {
	writePool *pgxpool.Pool
	readPool  *pgxpool.Pool
}

func NewClient(writePool, readPool *pgxpool.Pool) *Client {
	return &Client{writePool: writePool, readPool: readPool}
}

func (c *Client) Write(ctx context.Context) *pgxpool.Pool {
	return c.writePool
}

func (c *Client) Read(ctx context.Context) *pgxpool.Pool {
	return c.readPool
}
