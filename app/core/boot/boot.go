package cboot

import (
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

func StartListenTcpService() {
	listener, err := net.Listen(model.AppListenType, fmt.Sprintf("%s:%s", model.AppListenHost, model.AppListenPort))
	if err != nil {
		panic(err)
	}

	log.WithFields(log.Fields{"addr": listener.Addr().String()}).Info("StartListenTcpService")

	http.HandleFunc("/check-health", httpHandler)
	panic(http.Serve(listener, nil))
}

func RunAppPost() {
	time.Sleep(time.Hour)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, r.URL)
}
