package utils

import (
	"strings"
)

type MdadmState struct {
	Name   string `json:"n"`
	State  string `json:"s"`
	Level  string `json:"l"`
	Failed int    `json:"f"`
	Spare  int    `json:"x"`
	Total  int    `json:"t"`
}

func GetMdadmStates() (states []*MdadmState, err error) {
	states = []*MdadmState{}

	lines, err := ReadLines("/proc/mdstat")
	if err != nil {
		return
	}

	if len(lines) < 3 {
		return
	}

	for _, line := range lines {
		if strings.TrimSpace(line) == "" || line[0] == ' ' ||
			strings.HasPrefix(line, "Personalities") ||
			strings.HasPrefix(line, "unused") {

			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}

		state := &MdadmState{
			Name:   fields[0],
			State:  fields[2],
			Level:  fields[3],
			Failed: strings.Count(line, "(F)"),
			Spare:  strings.Count(line, "(S)"),
			Total:  strings.Count(line, "["),
		}

		states = append(states, state)
	}

	return
}
