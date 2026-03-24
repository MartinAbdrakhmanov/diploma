package ds

import "time"

var (
	DefaultTimeout = 2 * time.Second

	CleanupInterval   = 24 * time.Hour
	RetentionInterval = 365 * 24 * time.Hour
)

const (
	StatusTimeout = "timeout"
	StatusError   = "error"
	StatusSuccess = "success"
)

type Function struct {
	ID        string `db:"id"`
	UserId    string `db:"user_id"`
	Name      string `db:"name"`
	Image     string `db:"image"`
	Runtime   string `db:"runtime"`
	WasmPath  string `db:"wasm_path"`
	Timeout   int    `db:"timeout"`
	MaxMemory int64  `db:"max_memory"`
}

type ExecLog struct {
	ID string `db:"id"`

	FunctionID string `db:"function_id"`

	StartedAt  time.Time `db:"started_at"`
	FinishedAt time.Time `db:"finished_at"`
	DurationMs int64     `db:"duration_ms"`
	InitTimeMs int64     `db:"init_time_ms"`
	ExecTimeMs int64     `db:"exec_time_ms"`

	Status string `db:"status"`

	ExitCode     uint32 `db:"exit_code"`
	ErrorCode    string `db:"error_code"`
	ErrorMessage string `db:"error_message"`

	MaxMemoryBytes uint64 `db:"max_memory_bytes"`
	CPUTimeMs      uint64 `db:"cpu_time_ns"`

	CreatedAt time.Time `db:"created_at"`
}

type FunctionStats struct {
	AvgMemory        float64 `json:"avg_memory_bytes"`
	P95Memory        float64 `json:"p95_memory_bytes"`
	AvgDuration      float64 `json:"avg_duration_ms"`
	SuccessRate      float64 `json:"success_rate_percent"`
	TotalInvocations int     `json:"total_invocations"`
}
