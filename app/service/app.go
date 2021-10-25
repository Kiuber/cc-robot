package service

import (
	"cc-robot/core/tool/mysql"
	cyaml "cc-robot/core/tool/yaml"
	"cc-robot/dao"
	mexc "cc-robot/extern"
	"cc-robot/model"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"strings"
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
		mexcAPIData := mexc.Symbols(ctx)
		for _, symbol := range mexcAPIData.Payload.(model.SupportSymbols).Symbols {
			symbolList := strings.Split(symbol, "_")
			exchangeSymbol := dao.ExchangeSymbol{ExchangeName: "mexc", Symbol: symbol, Symbol1: symbolList[0], Symbol2: symbolList[1]}

			mysql.DB(ctx).Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(&exchangeSymbol)
			log.Infof("symbol %s", symbol)
		}
		time.Sleep(time.Second * 2)
	}
}
