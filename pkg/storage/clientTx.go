package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ClientTx struct {
	writePool *pgxpool.Pool
	readPool  *pgxpool.Pool
}

type Mode string

const (
	ReadMode  Mode = "read"
	WriteMode Mode = "write"
)

func NewClientTx(writePool, readPool *pgxpool.Pool) *ClientTx {
	return &ClientTx{writePool: writePool, readPool: readPool}
}

func (c *ClientTx) getConn(md Mode) *pgxpool.Pool {
	if md == ReadMode {
		return c.readPool
	}
	return c.writePool
}

func (c *ClientTx) logQuery(query string, mode Mode, args ...any) {
	log.Printf("Query: %s, in mode %s, with args %s \n", query, mode, args)
}

type txKey struct{}

// injectTx injects transaction to context
func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// extractTx extracts transaction from context
func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// WithinTransaction runs function within transaction
//
// The transaction commits when function were finished without error
func (c *ClientTx) WithinTransaction(ctx context.Context, mode Mode, tFunc func(ctx context.Context) error) error {
	// begin transaction
	conn := c.getConn(mode)
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// run callback
	err = tFunc(injectTx(ctx, tx))
	if err != nil {
		// if error, rollback
		tx.Rollback(ctx)
		return err
	}

	// if no error, commit
	tx.Commit(ctx)
	return nil
}

// Exec выполняет заданный SQL запрос с параметрами, возвращает ошибку или тэг запроса
func (c *ClientTx) Exec(ctx context.Context, mode Mode, query string, args ...any) (pgconn.CommandTag, error) {
	tx := extractTx(ctx)
	c.logQuery(query, mode, args...)
	if tx != nil {
		return tx.Exec(ctx, query, args...)
	}

	return c.getConn(mode).Exec(ctx, query, args...)
}

// Query выполняет запрос и возвращает объект для получения многострочного результата
func (c *ClientTx) Query(ctx context.Context, mode Mode, query string, args ...any) (pgx.Rows, error) {
	tx := extractTx(ctx)
	c.logQuery(query, mode, args...)
	if tx != nil {
		return tx.Query(ctx, query, args...)
	}

	return c.getConn(mode).Query(ctx, query, args...)
}

// QueryRow выполняет запрос и возвращает объект для получения одиночной строки
func (c *ClientTx) QueryRow(ctx context.Context, mode Mode, query string, args ...any) pgx.Row {
	tx := extractTx(ctx)
	c.logQuery(query, mode, args...)
	if tx != nil {
		return tx.QueryRow(ctx, query, args...)
	}

	return c.getConn(mode).QueryRow(ctx, query, args...)
}
