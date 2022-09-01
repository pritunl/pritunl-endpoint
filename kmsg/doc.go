package kmsg

import (
	"time"
)

const (
	Type = "kmsg"
)

type Kmsg struct {
	Timestamp time.Time `json:"t"`

	Boot     int64  `json:"b"`
	Priortiy int    `json:"p"`
	Sequence int64  `json:"s"`
	Message  string `json:"m"`
}

func (d *Kmsg) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *Kmsg) SetTimestamp(timestamp time.Time) {
	_ = timestamp
}

func (d *Kmsg) GetType() string {
	return Type
}
