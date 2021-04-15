package input

import (
	"time"

	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/sirupsen/logrus"
)

type Input struct {
	Name     string
	Rate     time.Duration
	Handler  func(stream *stream.Stream) error
	timstamp time.Time
}

func Run() {
	strm := stream.New()
	go strm.Run()

	for {
		for _, in := range inputs {
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
