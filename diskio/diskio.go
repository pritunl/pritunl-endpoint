package diskio

import (
	"strings"
	"time"
	"unicode"

	"github.com/dropbox/godropbox/errors"
	"github.com/pritunl/pritunl-endpoint/errortypes"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/stream"
	"github.com/shirou/gopsutil/disk"
)

var (
	prev map[string]disk.IOCountersStat
)

func Handler(stream *stream.Stream) (err error) {
	stats, err := disk.IOCounters()
	if err != nil {
		err = &errortypes.ParseError{
			errors.Wrap(err, "diskio: Failed to get disk io"),
		}
		return
	}

	statsMap := map[string]disk.IOCountersStat{}

	doc := &DiskIo{
		Disks: []*Disk{},
	}

	for _, stat := range stats {
		if strings.HasPrefix(stat.Name, "dm") {
			continue
		} else if strings.HasPrefix(stat.Name, "loop") {
			continue
		} else if strings.HasPrefix(stat.Name, "nvme") {
			l := len(stat.Name)
			if stat.Name[l-2] == 'p' || stat.Name[l-3] == 'p' {
				continue
			}
		} else if strings.HasPrefix(stat.Name, "md") {
		} else if strings.HasPrefix(stat.Name, "zram") {
		} else {
			l := len(stat.Name)
			if unicode.IsDigit(rune(stat.Name[l-1])) {
				continue
			}
		}

		statsMap[stat.Name] = stat

		prevStat, ok := prev[stat.Name]
		if ok {
			if stat.ReadBytes < prevStat.ReadBytes ||
				stat.WriteBytes < prevStat.WriteBytes ||
				stat.ReadCount < prevStat.ReadCount ||
				stat.WriteCount < prevStat.WriteCount ||
				stat.ReadTime < prevStat.ReadTime ||
				stat.WriteTime < prevStat.WriteTime ||
				stat.IoTime < prevStat.IoTime {

				ignore = true
			} else {
				doc.Disks = append(doc.Disks, &Disk{
					Name:       stat.Name,
					BytesRead:  stat.ReadBytes - prevStat.ReadBytes,
					BytesWrite: stat.WriteBytes - prevStat.WriteBytes,
					CountRead:  stat.ReadCount - prevStat.ReadCount,
					CountWrite: stat.WriteCount - prevStat.WriteCount,
					TimeRead:   stat.ReadTime - prevStat.ReadTime,
					TimeWrite:  stat.WriteTime - prevStat.WriteTime,
					TimeIo:     stat.IoTime - prevStat.IoTime,
				})
			}
		}
	}

	prev = statsMap

	if !ignore && len(doc.Disks) != 0 {
		stream.Append(doc)
	}

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
