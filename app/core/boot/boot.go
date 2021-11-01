package cboot

import (
	cjson "cc-robot/core/tool/json"
	cyaml "cc-robot/core/tool/yaml"
	"cc-robot/model"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
	"time"
)

var GV model.GlobalVariable
var env *string
var isDev bool

func PrepareCmdArgs() {
	env = flag.String("env", "dev", "runtime environment")
	flag.Parse()
}

func Init() {
	initEnv()
	initLog()
	initGV()
}

func initEnv() {
	isDev = *env == model.EnvDev
}

func initLog() {
	// log.SetReportCaller(true)
	formatter := &log.TextFormatter{
		FullTimestamp: true,
		TimestampFormat: time.RFC3339Nano,
	}
	log.SetFormatter(formatter)

	logLevel := log.InfoLevel
	if isDev {
		logLevel = log.DebugLevel
	}
	log.SetLevel(logLevel)
}

func initGV() {
	gv := &model.GlobalVariable{
		Env: *env,
		IsDev: isDev,
	}

	infra := &model.Infra{}
	api := &model.Api{}
	cyaml.LoadConfig("infra.yaml", infra)
	cyaml.LoadConfig("api.yaml", api)
	gv.Config.Infra = *infra
	gv.Config.Api = *api
	GV = *gv

	log.WithFields(log.Fields{"global variable": GV}).Info("initGV")
}

func RunAppPost() {
	time.Sleep(time.Hour)
}

func StartMockListenTcpService() {
	listener, err := net.Listen(model.MockListenType, fmt.Sprintf("%s:%s", model.MockListenHost, model.MockListenPort))
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("StartListenTcpService")

	for apiPath := range getMockData() {
		http.HandleFunc(apiPath, func(writer http.ResponseWriter, request *http.Request) {
			realMockData := map[string]interface{}{}
			cjson.UnmarshalFromFile("mock/mexc.json", &realMockData)

			for apiPath2, data  := range getMockData() {
				if apiPath2 == request.URL.Path {
					fmt.Fprintln(writer, cjson.Pretty(data))
				}
			}
		})
	}

	http.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		for apiPath := range getMockData() {
			fmt.Fprintln(writer, apiPath)
		}
	})
	go http.Serve(listener, nil)
}

func getMockData() map[string]interface{} {
	mockData := map[string]interface{}{}
	cjson.UnmarshalFromFile("mock/mexc.json", &mockData)
	return mockData
}