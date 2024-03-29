package check

import (
	"time"
)

const (
	Type = "check"
)

type Check struct {
	Timestamp time.Time `json:"t"`

	CheckId string   `json:"c"`
	Targets []string `json:"x"`
	Latency []int    `json:"l"`
	Errors  []string `json:"r"`
}

func (d *Check) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *Check) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *Check) GetType() string {
	return Type
}
