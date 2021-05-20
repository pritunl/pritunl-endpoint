package load

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/v3/load"
)

func Handler(stream *stream.Stream) (err error) {
	avgStat, err := load.Avg()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "load: Failed to get load average"),
		}
		return
	}

	doc := &Load{
		Load1:  avgStat.Load1,
		Load5:  avgStat.Load5,
		Load15: avgStat.Load15,
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
