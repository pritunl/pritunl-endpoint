package network

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/net"
)

func Handler(stream *stream.Stream) (err error) {
	stats, err := net.IOCounters(true)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "network: Failed to get network average"),
		}
		return
	}

	doc := &Network{
		Interfaces: []*Interface{},
	}

	for _, stat := range stats {
		doc.Interfaces = append(doc.Interfaces, &Interface{
			Name:        stat.Name,
			BytesSent:   stat.BytesSent,
			BytesRecv:   stat.BytesRecv,
			PacketsSent: stat.PacketsSent,
			PacketsRecv: stat.PacketsRecv,
			ErrorsSent:  stat.Errout,
			ErrorsRecv:  stat.Errin,
			DropsSent:   stat.Dropout,
			DropsRecv:   stat.Dropin,
			FifoSent:    stat.Fifoout,
			FifoRecv:    stat.Fifoin,
		})
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
