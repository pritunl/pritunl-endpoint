package disk

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/v3/disk"
)

var (
	ignoreTypesDefault = []string{
		"devtmpfs",
		"devfs",
		"overlay",
		"aufs",
		"squashfs",
	}
)

func Handler(stream *stream.Stream) (err error) {
	ignoreTypes := ignoreTypesDefault
	confIgnoreTypes := config.Config.Disk.IgnoreTypes
	if confIgnoreTypes != nil {
		ignoreTypes = confIgnoreTypes
	}

	ignoreTypesSet := set.NewSet()
	for _, ignoreType := range ignoreTypes {
		ignoreTypesSet.Add(ignoreType)
	}

	ignorePaths := []string{}
	confIgnorePaths := config.Config.Disk.IgnorePaths
	if confIgnorePaths != nil {
		ignorePaths = confIgnorePaths
	}

	ignorePathsSet := set.NewSet()
	for _, ignoreType := range ignorePaths {
		ignorePathsSet.Add(ignoreType)
	}

	parts, err := disk.Partitions(false)
	if err != nil {
		err = &errortypes.ReadError{
			errors.Wrap(err, "load: Failed to get disk partitions"),
		}
		return
	}

	mountpoints := []string{}
	for _, part := range parts {
		if ignoreTypesSet.Contains(part.Fstype) ||
			ignorePathsSet.Contains(part.Mountpoint) {

			continue
		}

		mountpoints = append(mountpoints, part.Mountpoint)
	}

	doc := &Disk{
		Mounts: []*Mount{},
	}

	for _, mountpoint := range mountpoints {
		usage, e := disk.Usage(mountpoint)
		if e != nil {
			continue
		}

		mount := &Mount{
			Path:   usage.Path,
			Format: usage.Fstype,
			Size:   usage.Total,
			Used:   usage.UsedPercent,
		}

		doc.Mounts = append(doc.Mounts, mount)
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
