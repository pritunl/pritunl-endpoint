package network

import (
	"time"
)

const (
	Type = "network"
)

type Interface struct {
	Name        string `json:"n"`
	BytesSent   uint64 `json:"bs"`
	BytesRecv   uint64 `json:"br"`
	PacketsSent uint64 `json:"ps"`
	PacketsRecv uint64 `json:"pr"`
	ErrorsSent  uint64 `json:"es"`
	ErrorsRecv  uint64 `json:"er"`
	DropsSent   uint64 `json:"ds"`
	DropsRecv   uint64 `json:"dr"`
	FifoSent    uint64 `json:"fs"`
	FifoRecv    uint64 `json:"fr"`
}

type Network struct {
	Timestamp time.Time `json:"t"`

	Interfaces []*Interface `json:"i"`
}

func (d *Network) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *Network) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *Network) GetType() string {
	return Type
}
