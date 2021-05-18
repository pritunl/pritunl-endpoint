package system

import (
	"time"
)

const (
	Type = "system"
)

type System struct {
	Timestamp time.Time `json:"t"`

	CpuUsage  float64 `json:"cu"`
	MemTotal  int     `json:"mt"`
	MemUsage  float64 `json:"mu"`
	SwapTotal int     `json:"st"`
	SwapUsage float64 `json:"su"`
}

func (d *System) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *System) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *System) GetType() string {
	return "system"
}
