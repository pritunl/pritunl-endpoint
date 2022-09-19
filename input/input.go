package input

import (
	"time"

	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/sirupsen/logrus"
)

type Input struct {
	Name     string
	Rate     time.Duration
	Startup  func(stream *stream.Stream) error
	Handler  func(stream *stream.Stream) error
	timstamp time.Time
}

func Run() {
	strm := stream.New()
	go strm.Run()

	for _, in := range inputs {
		if in.Startup != nil {
			err := in.Startup(strm)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("input: Input startup error")
			}
		}
	}

	for {
		for _, in := range inputs {
			if in.Handler == nil {
				continue
			}

			if time.Since(in.timstamp) > in.Rate {
				in.timstamp = time.Now()

				err := in.Handler(strm)
				if err != nil {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("input: Input handler error")
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
}
