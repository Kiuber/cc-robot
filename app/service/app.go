package service

import (
	cyaml "cc-robot/core/tool/yaml"
	mexc "cc-robot/extern"
	"cc-robot/module"
	log "github.com/sirupsen/logrus"
	"time"
)

func RunApp(ctx *module.Context) {
	updateCtx(ctx)
	initLogic(ctx)
}

func updateCtx(ctx *module.Context) {
	ctx.ApiConfig = cyaml.LoadConfig()
	log.WithFields(log.Fields{
		"cyaml.LoadConfig()": cyaml.LoadConfig(),
		"ctx": ctx,
	}).Info("updateCtx")
}

func initLogic(ctx *module.Context) {
	log.WithFields(log.Fields{
		"cyaml.LoadConfig()": cyaml.LoadConfig(),
		"ctx": ctx,
	}).Info("initLogic")

	for {
		log.Info("RunApp")
		mexc.AccountInfo(*ctx)
		time.Sleep(time.Second * 6)
	}
}