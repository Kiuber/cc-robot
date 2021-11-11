package main

import (
	cboot "cc-robot/core/boot"
	cjson "cc-robot/core/tool/json"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"cc-robot/service"
	"fmt"
	"go.uber.org/zap"
	"math/big"
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
	http.HandleFunc("/appear-symbol-pair", func(writer http.ResponseWriter, request *http.Request) {
		symbolPairParam := request.URL.Query().Get("symbol_pair")
		symbol1And2 := strings.Split(symbolPairParam, "_")
		appearSymbolPair := model.AppearSymbolPair{SymbolPair: symbolPairParam, Symbol1And2: symbol1And2}

		app.AcquireBetterPriceCh <- appearSymbolPair
		app.AppearSymbolPairManager[symbolPairParam] = model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair}

		fmt.Fprintln(writer, cjson.Pretty(buildSummary(app)))
	})
	http.HandleFunc("/update-symbol-pair-conf", func(writer http.ResponseWriter, request *http.Request) {
		symbolPairParam := request.URL.Query().Get("symbol_pair")
		expectedProfitRateParam := request.URL.Query().Get("expected_profit_rate")
		expectedProfitRate, ok := big.NewFloat(0).SetString(expectedProfitRateParam)
		if !ok {
			fmt.Fprintln(writer, "expectedProfitRateParam illegal")
			return
		}
		app.SymbolPairConf[symbolPairParam] = model.SymbolPairConf{
			BidCost:            app.SymbolPairConf[symbolPairParam].BidCost,
			ExpectedProfitRate: expectedProfitRate,
		}
		fmt.Fprintln(writer, cjson.Pretty(buildSummary(app)))
	})
	panic(http.Serve(listener, nil))
}

func buildSummary(app *service.App) map[string]interface{} {
	return map[string]interface{}{
		"AppearSymbolPairManager": app.AppearSymbolPairManager,
		"ListeningSymbolPair":     app.ListeningSymbolPair,
		"SymbolPairConf":          app.SymbolPairConf,
	}
}
