package load

import (
	"time"
)

const (
	Type = "load"
)

type Load struct {
	Timestamp time.Time `json:"t"`

	Load1  float64 `json:"lx"`
	Load5  float64 `json:"ly"`
	Load15 float64 `json:"lz"`
}

func (d *Load) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *Load) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *Load) GetType() string {
	return Type
}
