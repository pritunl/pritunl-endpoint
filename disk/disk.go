package disk

import (
	"time"

	"github.com/dropbox/godropbox/container/set"
	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/disk"
)

var (
	ignoreDefault = []string{
		"devtmpfs",
		"devfs",
		"overlay",
		"aufs",
		"squashfs",
	}
)

func Handler(stream *stream.Stream) (err error) {
	ignores := ignoreDefault
	confIgnores := config.Config.Disk.Ignores
	if confIgnores != nil {
		ignores = confIgnores
	}

	ignoresSet := set.NewSet()
	for _, ignore := range ignores {
		ignoresSet.Add(ignore)
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
		if ignoresSet.Contains(part.Fstype) {
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
