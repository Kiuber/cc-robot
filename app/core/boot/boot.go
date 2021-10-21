package cboot

import (
	"cc-robot/module"
	"cc-robot/service"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)

const (
	ListenHost  = "0"
	ListenPort = "3333"
	ListenType = "tcp"
)

func Init() {
	initLog()
	ctx := initContext()

	service.RunApp(ctx)

	startListenTcpService()
}

func initLog() {
	log.SetReportCaller(false)
}

func initContext() *module.Context {
	return &module.Context{
		IsDev: true,
	}
}

func startListenTcpService() {
	listener, err := net.Listen(ListenType, fmt.Sprintf("%s:%s", ListenHost, ListenPort))
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("listening on")

	http.HandleFunc("/check-health", httpHandler)
	panic(http.Serve(listener, nil))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, r.URL)
}
