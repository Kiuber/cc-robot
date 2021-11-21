package main

import (
	cboot "cc-robot/core/boot"
	cjson "cc-robot/core/tool/json"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"cc-robot/service"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
)

func main() {
	cboot.PrepareCmdArgs()
	cboot.Init()

	if cboot.GV.IsDev {
		cboot.StartMockListenTcpService()
	}

	app := service.RunApp()
	go service.RunCron(app)
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

	clog.EventLog.With(zap.String("addr", listener.Addr().String())).Info("StartListenTcpService")

	http.HandleFunc("/check-health", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, request.URL)
	})
	http.HandleFunc("/summary", func(writer http.ResponseWriter, request *http.Request) {
		fmt.Fprintln(writer, cjson.Pretty(buildSummary(app)))
	})
	http.HandleFunc("/check-support-symbol-pair", func(writer http.ResponseWriter, request *http.Request) {
		app.Exchange.SaveAPISupportSymbolPairsOfAllExchanges()
		app.Prime.CheckAndAlarmSymbolPairsOfAllExchanges()
		fmt.Fprintln(writer, cjson.Pretty(buildSummary(app)))
	})
	http.HandleFunc("/fetch-prime-symbol-pair", func(writer http.ResponseWriter, request *http.Request) {
		app.Prime.TryUpdatePrimeSymbolPair()
		fmt.Fprintln(writer, cjson.Pretty(buildSummary(app)))
	})
	panic(http.Serve(listener, nil))
}

func buildSummary(app *service.App) map[string]interface{} {
	return map[string]interface{}{
		"AppearSymbolPairManager": app.Prime.AppearSymbolPairManager,
		"ListeningSymbolPair":     app.Prime.ListeningSymbolPair,
		"SymbolPairConf":          app.Prime.SymbolPairConf,
	}
}
