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

func GetRepoQueryPath() (pth string, err error) {
	exists, err := utils.ExistsFile("/usr/bin/repoquery")
	if err != nil {
		return
	}

	if exists {
		pth = "/usr/bin/repoquery"
		return
	}

	return
}

func CheckUpdate() (count int, err error) {
	rqPth, err := GetRepoQueryPath()
	if err != nil {
		return
	}

	if rqPth == "" {
		return
	}

	output, exitCode, err := utils.ExecOutputCode(
		rqPth, "--upgrades", "--quiet")
	if err != nil {
		return
	}

	if exitCode != 0 {
		err = &errortypes.ExecError{
			errors.Newf(
				"dnf: Bad exit code %d from repoquery check update",
				exitCode,
			),
		}
		return
	}

	count = 0
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.Contains(line, ".src") {
			continue
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

	pth, _ := GetRepoQueryPath()

	if pth != "" {
		isDnf = true
		isDnfCached = true
	} else {
		isDnf = false
		isDnfCached = true
	}
	return isDnf
}
