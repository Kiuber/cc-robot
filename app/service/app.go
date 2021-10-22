package service

import (
	cyaml "cc-robot/core/tool/yaml"
	"cc-robot/model"
	log "github.com/sirupsen/logrus"
	"time"
)

func RunApp(ctx *model.Context) {
	updateCtx(ctx)
	initLogic(*ctx)
}

func updateCtx(ctx *model.Context) {
	apiConfig := &model.Api{}
	cyaml.LoadConfig("api.yaml", apiConfig)
	ctx.Config.Api = *apiConfig
}

func initLogic(ctx model.Context) {
	log.WithFields(log.Fields{"ctx": ctx}).Info("initLogic")

	for {
		log.Info("RunApp")
		// mexc.AccountInfo(ctx)
		// mysql.DB(ctx)
		time.Sleep(time.Second * 2)
	}
}
