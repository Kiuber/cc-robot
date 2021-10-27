package main

import (
	cboot "cc-robot/core/boot"
	"cc-robot/service"
)

func main() {
	cboot.Init()
	service.RunApp()
	cboot.StartListenTcpService()
}
