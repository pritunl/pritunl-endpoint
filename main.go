package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/pritunl/pritunl-endpoint/config"
	"github.com/pritunl/pritunl-endpoint/constants"
	"github.com/pritunl/pritunl-endpoint/disk"
	"github.com/pritunl/pritunl-endpoint/diskio"
	"github.com/pritunl/pritunl-endpoint/endpoint"
	"github.com/pritunl/pritunl-endpoint/input"
	"github.com/pritunl/pritunl-endpoint/load"
	"github.com/pritunl/pritunl-endpoint/logger"
	"github.com/pritunl/pritunl-endpoint/network"
	"github.com/pritunl/pritunl-endpoint/system"
)

const help = `
Usage: pritunl-endpoint COMMAND

Commands:
  version   Show version
  start     Start endpoint service
  register  Get default administrator password
`

func main() {
	defer time.Sleep(500 * time.Millisecond)

	flag.Usage = func() {
		fmt.Println(help)
	}

	flag.Parse()

	logger.Init()

	switch flag.Arg(0) {
	case "start":
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
		diskio.Register()
		network.Register()

		input.Run()

		return
	case "version":
		fmt.Printf("pritunl-endpoint v%s\n", constants.Version)
		return
	case "register":
		err := config.Init()
		if err != nil {
			panic(err)
		}

		err = endpoint.RegisterCmd()
		if err != nil {
			panic(err)
		}

		return
	}

	fmt.Println(help)
}
