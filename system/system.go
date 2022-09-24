package system

import (
	"fmt"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/constants"
	"github.com/pritunl/pritunl-endpoint/dnf"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/pritunl/pritunl-endpoint/utils"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
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
	hugeTotal int, hugeUsage float64, swapTotal int, swapUsage float64,
	err error) {

	m, err := utils.GetMemInfo()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "system: Failed to get memory usage"),
		}
		return
	}

	memTotal = int(m.Total / 1024)
	memUsage = m.UsedPercent
	swapTotal = int(m.SwapTotal / 1024)
	if m.SwapTotal != 0 {
		swapUsage = 100 * (1 - float64(
			m.SwapFree)/float64(m.SwapTotal))
	}

	hugeTotal = int(m.HugePagesTotal * m.HugePageSize / 1024)
	hugeUsage = m.HugePagesUsedPercent

	return
}

func Handler(stream *stream.Stream) (err error) {
	cpuCores, cpuUsage, err := getCpu()
	if err != nil {
		return
	}

	mTotal, mUsage, hTotal, hUsage, sTotal, sUsage, err := getMem()
	if err != nil {
		return
	}

	mdadm, err := utils.GetMdadmStates()
	if err != nil {
		return
	}

	if cpuUsage == 0 {
		return
	}

	info, err := host.Info()
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "system: Failed to get host info"),
		}
		return
	}

	virt := ""
	if info.VirtualizationRole == "guest" {
		virt = info.VirtualizationSystem
	}

	dnfCount := 0
	if dnf.IsDnf() {
		dnfCount, err = dnf.CheckUpdateCached()
		if err != nil {
			return
		}
	}

	doc := &System{
		Version:        constants.Version,
		Hostname:       info.Hostname,
		Uptime:         info.Uptime,
		Virtualization: virt,
		Platform: fmt.Sprintf("%s-%s-%s", info.OS,
			info.Platform, info.PlatformVersion),
		PackageUpdates: dnfCount,
		CpuCores:       cpuCores,
		CpuUsage:       cpuUsage,
		MemTotal:       mTotal,
		MemUsage:       mUsage,
		HugeTotal:      hTotal,
		HugeUsage:      hUsage,
		SwapTotal:      sTotal,
		SwapUsage:      sUsage,
		Mdadm:          mdadm,
	}

	stream.Append(doc)

	return
}

func Register() {
	in := &input.Input{
		Name:    Type,
		Rate:    60 * time.Second,
		Handler: Handler,
	}

	input.Register(in)
}
