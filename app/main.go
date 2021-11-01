package main

import (
	cboot "cc-robot/core/boot"
	"cc-robot/service"
)

func main() {
	cboot.PrepareCmdArgs()
	cboot.Init()
	go cboot.StartListenTcpService()
	service.RunApp()

	if cboot.GV.IsDev {
		cboot.RunAppPost()
	}
}
