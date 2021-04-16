package system

import (
	"fmt"
	"time"

	"github.com/pritunl/mongo-go-driver/bson/primitive"
)

const (
	Type = "system"
)

type System struct {
	Id        primitive.ObjectID `json:"i"`
	Timestamp time.Time          `json:"t"`
	Type      string             `json:"x"`

	CpuUsage  float64 `json:"cu"`
	MemTotal  int     `json:"mt"`
	MemUsage  float64 `json:"mu"`
	SwapTotal int     `json:"st"`
	SwapUsage float64 `json:"su"`
}

func (d *System) GetId() primitive.ObjectID {
	return d.Id
}

func (d *System) SetId(id primitive.ObjectID) {
	d.Id = id
}

func (d *System) GetTimestamp() time.Time {
	return d.Timestamp
}

func (d *System) SetTimestamp(timestamp time.Time) {
	d.Timestamp = timestamp
}

func (d *System) GetType() string {
	return d.Type
}

func (d *System) Print() {
	fmt.Println("***************************************************")
	fmt.Println("Id:", d.GetId().String())
	fmt.Println("Timestamp:", d.GetTimestamp())
	fmt.Println("Type:", d.GetType())

	fmt.Println("CpuUsage:", d.CpuUsage)
	fmt.Println("MemTotal:", d.MemTotal)
	fmt.Println("MemUsage:", d.MemUsage)
	fmt.Println("SwapTotal:", d.SwapTotal)
	fmt.Println("SwapUsage:", d.SwapUsage)
	fmt.Println("***************************************************")
}
