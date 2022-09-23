package stream

import (
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/utils"
)

var (
	checkLast   = map[string]time.Time{}
	CurrentConf = &Conf{}
)

type Check struct {
	Id         string    `json:"id"`
	Name       string    `json:"name"`
	Roles      []string  `json:"roles"`
	Frequency  int       `json:"frequency"`
	Type       string    `json:"type"`
	Targets    []string  `json:"targets"`
	Timeout    int       `json:"timeout"`
	Method     string    `json:"method"`
	StatusCode int       `json:"status_code"`
	Headers    []*Header `json:"headers"`
}

func (c *Check) Validate() (err error) {
	if c.Id == "" {
		err = &errortypes.ParseError{
			errors.New("stream: Check ID is invalid"),
		}
		return
	}

	if c.Name == "" {
		c.Name = "unknown"
	}

	if c.Roles == nil {
		c.Roles = []string{}
	}

	if c.Frequency <= 5 {
		c.Frequency = 10
	}

	if c.Frequency > 3600 {
		c.Frequency = 3600
	}

	if c.Targets == nil {
		c.Targets = []string{}
	}

	switch c.Type {
	case "http":
		break
	case "ping":
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("stream: Check type (%s) is invalid", c.Type),
		}
		return
	}

	switch c.Method {
	case "GET":
		break
	case "HEAD":
		break
	case "POST":
		break
	case "PUT":
		break
	case "DELETE":
		break
	default:
		err = &errortypes.ParseError{
			errors.Newf("stream: Check method (%s) is invalid", c.Method),
		}
		return
	}

	if c.Headers == nil {
		c.Headers = []*Header{}
	}

	if c.StatusCode <= 0 || c.StatusCode > 900 {
		c.StatusCode = 200
	}

	for _, header := range c.Headers {
		header.Key = utils.FilterStr(header.Key, 256)
		header.Value = utils.FilterStr(header.Value, 2048)
	}

	if c.Timeout < 1 {
		c.Timeout = 5
	} else if c.Timeout > 30 {
		c.Timeout = 30
	}

	return
}

func (c *Check) Ready() bool {
	if time.Since(checkLast[c.Id]) > time.Duration(
		c.Frequency)*time.Second {

		return true
	}
	return false
}

func (c *Check) SetNotReady() {
	checkLast[c.Id] = time.Now()
}

type Header struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type Conf struct {
	Checks []*Check `json:"checks"`
}
