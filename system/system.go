package system

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
)

func getCpu() (cores int, usage float64, err error) {
	cores, err = cpu.Counts(true)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "system: Failed to get CPU cores"),
		}
		return
	}

	cpuUsages, err := cpu.Percent(0, false)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "system: Failed to get CPU percent"),
		}
		return
	}

	if len(cpuUsages) != 1 {
		err = &errortypes.ParseError{
			errors.New("system: Invalid CPU percent"),
		}
		return
	}

	usage = cpuUsages[0]

	return
}

func getMem() (memTotal int, memUsage float64,
	swapTotal int, swapUsage float64, err error) {

	memUsages, err := mem.VirtualMemory()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "system: Failed to get memory usage"),
		}
		return
	}

	memTotal = int(memUsages.Total / OneMebibyte)
	memUsage = memUsages.UsedPercent
	swapTotal = int(memUsages.SwapTotal / OneMebibyte)
	if memUsages.SwapTotal != 0 {
		swapUsage = 100 * (1 - float64(
			memUsages.SwapFree)/float64(memUsages.SwapTotal))
	}

	return
}

func Handler(stream *stream.Stream) (err error) {
	cpuCores, cpuUsage, err := getCpu()
	if err != nil {
		return
	}

	memTotal, memUsage, swapTotal, swapUsage, err := getMem()
	if err != nil {
		return
	}

	if cpuUsage == 0 {
		return
	}

	doc := &System{
		CpuCores:  cpuCores,
		CpuUsage:  cpuUsage,
		MemTotal:  memTotal,
		MemUsage:  memUsage,
		SwapTotal: swapTotal,
		SwapUsage: swapUsage,
	}

	stream.Append(doc)

	return
}

func Register() {
	in := &input.Input{
		Name:    "cpu",
		Rate:    60 * time.Second,
		Handler: Handler,
	}

	input.Register(in)
}
