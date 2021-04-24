package main

import (
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/logger"
	"github.com/pritunl/pritunl-endpoint/system"
)

func main() {
	logger.Init()

	system.Register()

	input.Run()
}
