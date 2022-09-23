package dnf

import (
	"strings"
	"time"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/utils"
)

var (
	lastUpdateCount int
	lastUpdateTime  time.Time
	isDnf           = false
	isDnfCached     = false
)

func GetDnfPath() (pth string, err error) {
	exists, err := utils.ExistsFile("/usr/bin/dnf")
	if err != nil {
		return
	}

	if exists {
		pth = "/usr/bin/dnf"
		return
	}

	exists, err = utils.ExistsFile("/usr/bin/yum")
	if err != nil {
		return
	}

	if exists {
		pth = "/usr/bin/yum"
		return
	}

	return
}

func CheckUpdate() (count int, err error) {
	dnfPth, err := GetDnfPath()
	if err != nil {
		return
	}

	if dnfPth == "" {
		return
	}

	output, exitCode, err := utils.ExecOutputCode(
		dnfPth, "check-update", "-q")
	if err != nil {
		return
	}

	if exitCode == 0 {
		return
	} else if exitCode != 100 {
		err = &errortypes.ExecError{
			errors.Newf(
				"dnf: Bad exit code %d from dnf check update",
				exitCode,
			),
		}
		return
	}

	count = 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "Obsoleting Packages") {
			break
		}

		count += 1
	}

	return
}

func CheckUpdateCached() (count int, err error) {
	if time.Since(lastUpdateTime) < 1*time.Hour {
		count = lastUpdateCount
		return
	}

	count, err = CheckUpdate()
	if err != nil {
		return
	}

	lastUpdateCount = count
	lastUpdateTime = time.Now()

	return
}

func IsDnf() bool {
	if isDnfCached {
		return isDnf
	}

	pth, _ := GetDnfPath()

	if pth != "" {
		isDnf = true
		isDnfCached = true
	} else {
		isDnf = false
		isDnfCached = true
	}
	return isDnf
}
