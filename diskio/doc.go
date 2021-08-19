package diskio

import (
	"time"
)

const (
	Type = "diskio"
)

type DiskIo struct {
	Timestamp time.Time `json:"t"`

	Disks []*Disk `json:"d"`
}

type Disk struct {
	Name       string `json:"n"`
	BytesRead  uint64 `json:"br"`
	BytesWrite uint64 `json:"bw"`
	CountRead  uint64 `json:"cr"`
	CountWrite uint64 `json:"cw"`
	TimeRead   uint64 `json:"tr"`
	TimeWrite  uint64 `json:"tw"`
	TimeIo     uint64 `json:"ti"`
}

func (d *DiskIo) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *DiskIo) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *DiskIo) GetType() string {
	return Type
}
