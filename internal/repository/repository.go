package repository

import (
	"context"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/MartinAbdrakhmanov/diploma/pkg/storage"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type dbManager interface {
	Exec(ctx context.Context, md storage.Mode, query string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, md storage.Mode, query string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, md storage.Mode, query string, args ...any) pgx.Row
	WithinTransaction(ctx context.Context, md storage.Mode, tFunc func(ctx context.Context) error) error
}

type Repository struct {
	db dbManager
}

func New(dbManager dbManager) *Repository {
	return &Repository{
		db: dbManager,
	}
}

func (r *Repository) SaveFunction(ctx context.Context, function ds.Function) (id string, err error) {
	query := `
	INSERT INTO functions (user_id, "name", runtime, wasm_path, "image", "timeout", max_memory)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
	ON CONFLICT (user_id, "name")
    DO UPDATE SET
        updated_at = CURRENT_TIMESTAMP
	RETURNING id;
	`
	err = r.db.QueryRow(ctx, storage.WriteMode, query,
		function.UserId,
		function.Name,
		function.Runtime,
		function.WasmPath,
		function.Image,
		function.Timeout,
		function.MaxMemory,
	).Scan(&id)

	if err != nil {
		return "", nil
	}

	return id, nil
}
