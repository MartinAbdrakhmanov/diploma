package ds

import "time"

var (
	DefaultTimeout = 2 * time.Second
)

const (
	StatusTimeout = "timeout"
	StatusError   = "error"
	StatusSuccess = "success"
)

type Function struct {
	ID        string   `db:"id"`
	UserId    string   `db:"user_id"`
	Name      string   `db:"name"`
	Image     string   `db:"image"`
	Runtime   string   `db:"runtime"`
	WasmPath  string   `db:"wasm_path"`
	Timeout   int      `db:"timeout"`
	MaxMemory int64    `db:"max_memory"`
	Args      []string // is it even needed??
}

type ExecLog struct {
	ID string `db:"id"`

	FunctionID string `db:"function_id"`

	StartedAt  time.Time `db:"started_at"`
	FinishedAt time.Time `db:"finished_at"`
	DurationMs int64     `db:"duration_ms"`

	Status string `db:"status"`

	ExitCode     uint32 `db:"exit_code"`
	ErrorCode    string `db:"error_code"`
	ErrorMessage string `db:"error_message"`

	CreatedAt time.Time `db:"created_at"`
}
