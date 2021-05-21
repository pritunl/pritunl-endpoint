package disk

import (
	"time"
)

const (
	Type = "disk"
)

type Mount struct {
	Path   string  `json:"p"`
	Format string  `json:"f"`
	Size   uint64  `json:"s"`
	Used   float64 `json:"u"`
}

type Disk struct {
	Timestamp time.Time `json:"t"`

	Mounts []*Mount `json:"m"`
}

func (d *Disk) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *Disk) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *Disk) GetType() string {
	return Type
}
