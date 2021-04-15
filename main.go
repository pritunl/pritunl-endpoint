package main

import (
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/system"
)

func main() {
	system.Register()

	input.Run()
}
