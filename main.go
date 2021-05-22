package main

import (
	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/disk"
	"github.com/pritunl/pritunl-endpoint/endpoint"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/load"
	"github.com/pritunl/pritunl-endpoint/logger"
	"github.com/pritunl/pritunl-endpoint/network"
	"github.com/pritunl/pritunl-endpoint/system"
)

func main() {
	logger.Init()

	err := config.Init()
	if err != nil {
		panic(err)
	}

	err = endpoint.Init()
	if err != nil {
		panic(err)
	}

	system.Register()
	load.Register()
	disk.Register()
	network.Register()

	input.Run()
}
