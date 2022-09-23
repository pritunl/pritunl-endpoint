package check

import (
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/sirupsen/logrus"
)

type checker struct {
	stream *stream.Stream
}

func (c *checker) runCheckHttp(check *stream.Check, target string) (
	latency int, shortErr, err error) {

	timeout := time.Duration(check.Timeout) * time.Second

	dailer := &net.Dialer{
		Timeout:   timeout,
		KeepAlive: 30 * time.Second,
	}

	clientTransport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           dailer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Transport: clientTransport,
		Timeout:   timeout,
	}

	u, err := url.Parse(target)
	if err != nil {
		shortErr = err
		err = &errortypes.ParseError{
			errors.Wrap(err, "check: Failed to parse check URL"),
		}
		return
	}

	req, err := http.NewRequest(
		check.Method,
		u.String(),
		nil,
	)
	if err != nil {
		shortErr = err
		err = &errortypes.RequestError{
			errors.Wrap(err, "check: Request create error"),
		}
		return
	}

	if check.Headers != nil {
		for _, header := range check.Headers {
			req.Header.Set(header.Key, header.Value)
		}
	}

	start := time.Now()
	res, err := client.Do(req)
	latency = int(time.Since(start).Milliseconds())
	if err != nil {
		shortErr = err
		err = &errortypes.RequestError{
			errors.Wrap(err, "check: Request run error"),
		}
		return
	}
	defer res.Body.Close()

	if res.StatusCode != check.StatusCode {
		shortErr = err
		err = &errortypes.RequestError{
			errors.Wrapf(
				err,
				"check: Request status error %d",
				res.StatusCode,
			),
		}
		return
	}

	return
}

func (c *checker) runCheck(check *stream.Check) (err error) {
	switch check.Type {
	case "http":
		targets := []string{}
		latencies := []int{}
		errs := []string{}

		if check.Targets != nil {
			for _, target := range check.Targets {
				latency, shortErr, checkErr := c.runCheckHttp(check, target)
				checkErrStr := ""
				if shortErr != nil {
					latency = 0
					checkErrStr = shortErr.Error()
				}
				if checkErr != nil {
					logrus.WithFields(logrus.Fields{
						"error": checkErr,
					}).Error("check: Check run failed")
				}

				targets = append(targets, target)
				latencies = append(latencies, latency)
				errs = append(errs, checkErrStr)
			}
		}

		doc := &Check{
			CheckId: check.Id,
			Targets: targets,
			Latency: latencies,
			Errors:  errs,
		}

		c.stream.Append(doc)

		break
	case "ping":
		break
	default:
		logrus.WithFields(logrus.Fields{
			"type": check.Type,
		}).Warn("check: Ignoring unknown check type")
	}

	return
}

func (c *checker) Run(strm *stream.Stream) {
	c.stream = strm

	for {
		time.Sleep(1 * time.Second)

		conf := stream.CurrentConf
		if conf == nil || conf.Checks == nil || len(conf.Checks) == 0 {
			continue
		}

		for _, check := range conf.Checks {
			err := check.Validate()
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"error": err,
				}).Error("check: Check validate failed")
				continue
			}

			if check.Ready() {
				check.SetNotReady()

				go func(check *stream.Check) {
					err := c.runCheck(check)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"check_id": check.Id,
							"error":    err,
						}).Error("check: Failed to run check")
					}
				}(check)
			}
		}
	}
}

func startup(stream *stream.Stream) (err error) {
	checkr := &checker{}
	go checkr.Run(stream)

	return
}

func Register() {
	in := &input.Input{
		Name:    Type,
		Startup: startup,
	}

	input.Register(in)
}
