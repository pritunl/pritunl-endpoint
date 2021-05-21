package system

import (
	"time"
)

const (
	Type = "system"
)

type System struct {
	Timestamp time.Time `json:"t"`

	Hostname       string  `json:"h"`
	Uptime         uint64  `json:"u"`
	Virtualization string  `json:"v"`
	Platform       string  `json:"p"`
	Processes      uint64  `json:"pc"`
	CpuCores       int     `json:"cc"`
	CpuUsage       float64 `json:"cu"`
	MemTotal       int     `json:"mt"`
	MemUsage       float64 `json:"mu"`
	SwapTotal      int     `json:"st"`
	SwapUsage      float64 `json:"su"`
}

func (d *System) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *System) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *System) GetType() string {
	return Type
}
