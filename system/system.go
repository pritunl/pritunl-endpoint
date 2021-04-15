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

func getCpu() (usage float64, err error) {
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

	memTotal = int(memUsages.Total / 1048576)
	memUsage = memUsages.UsedPercent
	swapTotal = int(memUsages.SwapTotal / 1048576)
	if memUsages.SwapTotal != 0 {
		swapUsage = 100 * (1 - float64(
			memUsages.SwapFree)/float64(memUsages.SwapTotal))
	}

	return
}

func Handler(stream *stream.Stream) (err error) {
	cpuUsage, err := getCpu()
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

	fields := map[string]interface{}{
		"cpu_usage":  cpuUsage,
		"mem_total":  memTotal,
		"mem_usage":  memUsage,
		"swap_total": swapTotal,
		"swap_usage": swapUsage,
	}

	stream.Append("system", fields)

	return
}

func Register() {
	in := &input.Input{
		Name:    "cpu",
		Rate:    5 * time.Second,
		Handler: Handler,
	}

	input.Register(in)
}
