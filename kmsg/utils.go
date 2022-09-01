package kmsg

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/utils"
)

func GetBoottime() (boottime time.Time, err error) {
	output, err := utils.ExecOutput("uptime", "-s")
	if err != nil {
		return
	}

	boottime, err = time.Parse(
		"2006-01-02 15:04:05",
		strings.TrimSpace(output),
	)
	if err != nil {
		err = errortypes.ReadError{
			errors.Wrap(err, "kmsg: Failed to parse boottime"),
		}
		return
	}

	return
}
