package main

import (
	cboot "cc-robot/core/boot"
	cjson "cc-robot/core/tool/json"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"cc-robot/service"
	"fmt"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"strings"
)

func main() {
	cboot.PrepareCmdArgs()
	cboot.Init()

	if cboot.GV.IsDev {
		cboot.StartMockListenTcpService()
	}

	app := service.RunApp()
	startAppListenTcpService(app)

	if cboot.GV.IsDev {
		cboot.RunAppPost()
	}
}

func startAppListenTcpService(app *service.App) {
	listener, err := net.Listen(model.AppListenType, fmt.Sprintf("%s:%s", model.AppListenHost, model.AppListenPort))
	if err != nil {
		panic(err)
	}

	clog.EventLog().WithFields(logrus.Fields{"addr": listener.Addr().String()}).Info("StartListenTcpService")

	http.HandleFunc("/check-health", httpHandler)
	http.HandleFunc("/test-appear-symbol-pair", func(writer http.ResponseWriter, request *http.Request) {
		symbolPair := request.URL.Query().Get("symbol_pair")
		symbol1And2 := strings.Split(symbolPair, "_")
		appearSymbolPair := model.AppearSymbolPair{SymbolPair: symbolPair, Symbol1And2: symbol1And2}
		fmt.Fprintln(writer, cjson.Pretty(appearSymbolPair))
		app.BetterPriceCh <- appearSymbolPair
	})
	panic(http.Serve(listener, nil))
}

func httpHandler(writer http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(writer, r.URL)
}
