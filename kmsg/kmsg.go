package kmsg

import (
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/sirupsen/logrus"
)

type reader struct {
	stream   *stream.Stream
	boottime time.Time
}

func (r *reader) Handle(msgRaw string) (err error) {
	doc := &Kmsg{}

	msgs := strings.SplitN(msgRaw, ";", 2)
	if len(msgs) != 2 {
		return
	}

	meta := msgs[0]

	doc.Message = msgs[1]

	metas := strings.Split(meta, ",")
	if len(metas) < 3 {
		return
	}

	doc.Priortiy, err = strconv.Atoi(metas[0])
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "kmsg: Failed to parse doc priority"),
		}
		return
	}

	doc.Sequence, err = strconv.ParseInt(metas[1], 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "kmsg: Failed to parse doc sequence"),
		}
		return
	}

	timestamp, err := strconv.ParseInt(metas[2], 10, 64)
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "kmsg: Failed to parse doc timestamp"),
		}
		return
	}

	doc.Timestamp = r.boottime.Add(
		time.Duration(timestamp) * time.Microsecond)

	doc.Boot = r.boottime.Unix()

	r.stream.Append(doc)

	return
}

func (r *reader) Read() (err error) {
	file, err := os.Open("/dev/kmsg")
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "kmsg: Failed to open kmsg"),
		}
		return
	}
	defer func() {
		_ = file.Close()
	}()

	n := 0
	buffer := make([]byte, 4096)

	for {
		n, err = file.Read(buffer)
		if err != nil {
			if err == io.EOF || err == syscall.EPIPE {
				err = nil
				return
			}

			err = &errortypes.ReadError{
				errors.Wrap(err, "kmsg: Failed to read kmsg"),
			}
			return
		}

		err = r.Handle(string(buffer[:n]))
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("kmsg: Failed to handle kmsg")
			err = nil
		}
	}
}

func (r *reader) Run(strm *stream.Stream) {
	var err error

	r.stream = strm

	r.boottime, err = GetBoottime()
	if err != nil {
		panic(err)
	}

	for {
		err = r.Read()
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("kmsg: Input handler error")
		}

		time.Sleep(5 * time.Second)
	}

	return
}

func startup(stream *stream.Stream) (err error) {
	readr := &reader{}
	go readr.Run(stream)

	return
}

func Register() {
	in := &input.Input{
		Name:    Type,
		Startup: startup,
	}

	input.Register(in)
}
