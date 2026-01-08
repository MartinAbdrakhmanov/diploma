package ds

import "time"

type Function struct {
	ID        string    `db:"id"`
	UserId    string    `db:"user_id"`
	Name      string    `db:"name"`
	Image     string    `db:"image"`
	Runtime   string    `db:"runtime"`
	WasmPath  string    `db:"wasm_path"`
	Timeout   time.Time `db:"timeout"`
	MaxMemory int64     `db:"max_memory"`
	Args      []string  // is it even needed??
}
