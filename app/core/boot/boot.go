package cboot

import (
	cjson "cc-robot/core/tool/json"
	cyaml "cc-robot/core/tool/yaml"
	"cc-robot/model"
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
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
	log.SetReportCaller(true)
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

	files, err := ioutil.ReadDir("mock/")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		names := strings.Split(name, ".")
		basePath := fmt.Sprintf("/%s", names[0])

		for apiPath := range getMockData(name) {
			http.HandleFunc(fmt.Sprintf("%s%s", basePath, apiPath), func(writer http.ResponseWriter, request *http.Request) {
				time.Sleep(3 * time.Second)
				for apiPath2, data  := range getMockData(name) {
					if fmt.Sprintf("%s%s", basePath, apiPath2) == request.URL.Path {
						fmt.Fprintln(writer, cjson.Pretty(data))
					}
				}
			})
		}

		http.HandleFunc(fmt.Sprintf("%s", basePath), func(writer http.ResponseWriter, request *http.Request) {
			for apiPath := range getMockData(name) {
				fmt.Fprintln(writer, fmt.Sprintf("%s%s", basePath, apiPath))
			}
		})
	}

	log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("StartMockListenTcpService")
	go http.Serve(listener, nil)
}

func getMockData(name string) map[string]interface{} {
	mockData := map[string]interface{}{}
	cjson.UnmarshalFromFile(fmt.Sprintf("mock/%s", name), &mockData)
	return mockData
}