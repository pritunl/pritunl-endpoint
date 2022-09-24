package system

import (
	"time"

	"github.com/pritunl/pritunl-endpoint/utils"
)

const (
	Type = "system"
)

type System struct {
	Timestamp time.Time `json:"t"`

	Version        string              `json:"ev"`
	Hostname       string              `json:"h"`
	Uptime         uint64              `json:"u"`
	Virtualization string              `json:"v"`
	Platform       string              `json:"p"`
	PackageUpdates int                 `json:"pu"`
	Processes      uint64              `json:"pc"`
	CpuCores       int                 `json:"cc"`
	CpuUsage       float64             `json:"cu"`
	MemTotal       int                 `json:"mt"`
	MemUsage       float64             `json:"mu"`
	HugeTotal      int                 `json:"ht"`
	HugeUsage      float64             `json:"hu"`
	SwapTotal      int                 `json:"st"`
	SwapUsage      float64             `json:"su"`
	Mdadm          []*utils.MdadmState `json:"ra"`
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
