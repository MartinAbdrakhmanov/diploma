package repository

import (
	"context"
	"errors"

	"github.com/MartinAbdrakhmanov/diploma/internal/ds"
	"github.com/MartinAbdrakhmanov/diploma/pkg/storage"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5"
)

type Repository struct {
	db *storage.Client
}

func New(dbManager *storage.Client) *Repository {
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
	err = r.db.Write(ctx).QueryRow(ctx, query,
		function.UserId,
		function.Name,
		function.Runtime,
		function.WasmPath,
		function.Image,
		function.Timeout,
		function.MaxMemory,
	).Scan(&id)

	if err != nil {
		return "", err
	}

	return id, nil
}
func (r *Repository) GetFunction(ctx context.Context, userID, id string) (function ds.Function, err error) {
	query := `
	SELECT id, user_id, "name", runtime, wasm_path, "image", "timeout", max_memory
	FROM functions
	WHERE user_id = $1 AND id = $2
	`
	err = pgxscan.Get(ctx, r.db.Read(ctx), &function, query, userID, id)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ds.Function{}, nil
		}
		return ds.Function{}, err
	}

	return function, nil
}

func (r *Repository) SaveLog(ctx context.Context, log ds.ExecLog) error {
	query := `
	INSERT INTO execution_logs (function_id, started_at, finished_at, duration_ms, status, exit_code, error_code, error_message)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Write(ctx).Exec(ctx, query,
		log.FunctionID,
		log.StartedAt,
		log.FinishedAt,
		log.DurationMs,
		log.Status,
		log.ExitCode,
		log.ErrorCode,
		log.ErrorMessage,
	)

	return err
}

func (r *Repository) DeleteFunction(ctx context.Context, userID, id string) error {
	query := `
	DELETE FROM functions
	WHERE user_id = $1 AND id = $2
	`

	_, err := r.db.Write(ctx).Exec(ctx, query, userID, id)

	return err
}
