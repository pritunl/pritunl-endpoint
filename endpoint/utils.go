package endpoint

import (
	"flag"
	"fmt"
	"net/url"
	"strings"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/sirupsen/logrus"
)

func RegisterCmd() (err error) {
	uri := flag.Arg(1)

	if uri != "" {
		u, e := url.Parse(uri)
		if e != nil {
			err = &errortypes.ParseError{
				errors.Wrap(e, "endpoint: Failed to parse register uri"),
			}
			return
		}

		registerKey := u.Path
		if len(registerKey) < 32 {
			err = &errortypes.ParseError{
				errors.New("endpoint: Invalid register key in uri"),
			}
			return
		}

		registerKey = registerKey[1:]
		registerKeys := strings.Split(registerKey, "_")
		if len(registerKeys) != 2 {
			err = &errortypes.ParseError{
				errors.New("endpoint: Invalid register key in uri"),
			}
			return
		}

		config.Config.RemoteHosts = []string{u.Host}
		config.Config.Id = registerKeys[0]
		config.Config.Secret = registerKeys[1]
		config.Config.PublicKey = ""
		config.Config.PrivateKey = ""
		config.Config.ServerPublicKey = ""

		err = config.Save()
		if err != nil {
			return
		}
	} else {
		hostname := ""
		fmt.Print("Enter Pritunl Zero hostname: ")
		fmt.Scan(&hostname)

		if strings.HasPrefix(hostname, "https://") {
			u, e := url.Parse(hostname)
			if e != nil {
				err = &errortypes.ParseError{
					errors.Wrap(e, "endpoint: Failed to parse input"),
				}
				return
			}

			hostname = u.Host
		}

		if hostname == "" {
			err = &errortypes.ParseError{
				errors.New("endpoint: Invalid hostname"),
			}
			return
		}

		registerKey := ""
		fmt.Print("Enter registration key: ")
		fmt.Scan(&registerKey)

		if registerKey == "" {
			err = &errortypes.ParseError{
				errors.New("endpoint: Invalid registration key"),
			}
			return
		}

		registerKeys := strings.Split(registerKey, "_")
		if len(registerKeys) != 2 {
			err = &errortypes.ParseError{
				errors.New("endpoint: Invalid register key"),
			}
			return
		}

		config.Config.RemoteHosts = []string{hostname}
		config.Config.Id = registerKeys[0]
		config.Config.Secret = registerKeys[1]
		config.Config.PublicKey = ""
		config.Config.PrivateKey = ""
		config.Config.ServerPublicKey = ""

		err = config.Save()
		if err != nil {
			return
		}
	}

	logrus.WithFields(logrus.Fields{
		"endpoint_id":       config.Config.Id,
		"pritunl_zero_host": config.Config.RemoteHosts[0],
	}).Info("endpoint: Registration key saved")

	return
}
