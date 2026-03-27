package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// TODO rethink conflict resolution
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

func (r *Repository) FunctionInfo(ctx context.Context, userID, id string) (function ds.Function, err error) {
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
	INSERT INTO execution_logs (function_id, started_at, finished_at, duration_ms, status, exit_code, error_code, error_message, max_memory_bytes, cpu_time_ms, init_time_ms, exec_time_ms)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
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
		log.MaxMemoryBytes,
		log.CPUTimeMs,
		log.InitTimeMs,
		log.ExecTimeMs,
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

func (r *Repository) FunctionStats(ctx context.Context, userID, functionID string) (ds.FunctionStats, error) {
	var stats ds.FunctionStats

	query := `
    WITH last_calls AS (
        SELECT duration_ms, max_memory_bytes, status
        FROM execution_logs l
        JOIN functions f ON l.function_id = f.id
        WHERE f.user_id = $1 AND f.id = $2
        ORDER BY l.created_at DESC
        LIMIT 100
    )
    SELECT 
        COALESCE(AVG(max_memory_bytes), 0),
        COALESCE(percentile_cont(0.95) WITHIN GROUP (ORDER BY max_memory_bytes), 0),
        COALESCE(AVG(duration_ms), 0),
        COALESCE(COUNT(*) FILTER (WHERE status = 'success')::float / NULLIF(COUNT(*), 0) * 100, 0),
        COUNT(*)
    FROM last_calls
    `

	err := r.db.Read(ctx).QueryRow(ctx, query, userID, functionID).Scan(
		&stats.AvgMemory,
		&stats.P95Memory,
		&stats.AvgDuration,
		&stats.SuccessRate,
		&stats.TotalInvocations,
	)

	return stats, err
}

func (r *Repository) UpdateFunction(ctx context.Context, functionID string) error {
	query := `
	UPDATE functions 
	SET last_called_at = NOW() 
	WHERE id = $1`

	_, err := r.db.Write(ctx).Exec(ctx, query, functionID)

	return err
}

func (r *Repository) FunctionLastCalledAt(ctx context.Context, functionID string) (time.Time, error) {
	query := `
	SELECT last_called_at
	FROM functions
	WHERE id = $1
	`
	var lastCalledAt time.Time
	err := pgxscan.Get(ctx, r.db.Read(ctx), &lastCalledAt, query, functionID)

	if err != nil {
		return lastCalledAt, err
	}

	return lastCalledAt, nil
}

func (r *Repository) GetExpiredFunctions(ctx context.Context, retentionInterval time.Duration) ([]ds.Function, error) {
	var functions []ds.Function

	query := `
		SELECT id, user_id, "name", runtime, wasm_path, "image"
		FROM functions
		WHERE last_used_at < NOW() - $1::interval
	`
	err := pgxscan.Select(ctx, r.db.Read(ctx), &functions, query, retentionInterval.String())

	if err != nil {
		return nil, fmt.Errorf("failed to fetch expired functions: %w", err)
	}

	return functions, nil
}

func (r *Repository) UserFunctions(ctx context.Context, userID string) ([]ds.Function, error) {
	var functions []ds.Function
	query := `
	SELECT id, name, runtime FROM functions WHERE user_id = $1
	`

	err := pgxscan.Select(ctx, r.db.Read(ctx), &functions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user functions: %w", err)
	}

	return functions, nil
}
