package network

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/v3/net"
)

var (
	prev map[string]net.IOCountersStat
)

func Handler(stream *stream.Stream) (err error) {
	stats, err := net.IOCounters(true)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "network: Failed to get network average"),
		}
		return
	}

	statsMap := map[string]net.IOCountersStat{}

	doc := &Network{
		Interfaces: []*Interface{},
	}

	ignore := false
	for _, stat := range stats {
		statsMap[stat.Name] = stat

		prevStat, ok := prev[stat.Name]
		if ok {
			if stat.BytesSent < prevStat.BytesSent ||
				stat.BytesRecv < prevStat.BytesRecv ||
				stat.PacketsSent < prevStat.PacketsSent ||
				stat.PacketsRecv < prevStat.PacketsRecv ||
				stat.Errout < prevStat.Errout ||
				stat.Errin < prevStat.Errin ||
				stat.Dropout < prevStat.Dropout ||
				stat.Dropin < prevStat.Dropin ||
				stat.Fifoout < prevStat.Fifoout ||
				stat.Fifoin < prevStat.Fifoin {

				ignore = true
			} else {
				doc.Interfaces = append(doc.Interfaces, &Interface{
					Name:        stat.Name,
					BytesSent:   stat.BytesSent - prevStat.BytesSent,
					BytesRecv:   stat.BytesRecv - prevStat.BytesRecv,
					PacketsSent: stat.PacketsSent - prevStat.PacketsSent,
					PacketsRecv: stat.PacketsRecv - prevStat.PacketsRecv,
					ErrorsSent:  stat.Errout - prevStat.Errout,
					ErrorsRecv:  stat.Errin - prevStat.Errin,
					DropsSent:   stat.Dropout - prevStat.Dropout,
					DropsRecv:   stat.Dropin - prevStat.Dropin,
					FifoSent:    stat.Fifoout - prevStat.Fifoout,
					FifoRecv:    stat.Fifoin - prevStat.Fifoin,
				})
			}
		}
	}

	prev = statsMap

	if !ignore && len(doc.Interfaces) != 0 {
		stream.Append(doc)
	}

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
