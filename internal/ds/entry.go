package ds

import (
	"time"
)

type Entry struct {
	UserId    string
	Name      string
	Files     map[string][]byte
	Runtime   string
	Timeout   time.Time
	MaxMemory int64
}

func (e Entry) ToFunction() Function {
	return Function{
		UserId:    e.UserId,
		Name:      e.Name,
		Runtime:   e.Runtime,
		Timeout:   e.Timeout,
		MaxMemory: e.MaxMemory,
	}
}
