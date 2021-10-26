package service

import (
	"cc-robot/core/tool/mysql"
	"cc-robot/core/tool/redis"
	cyaml "cc-robot/core/tool/yaml"
	"cc-robot/dao"
	mexc "cc-robot/extern"
	"cc-robot/model"
	"context"
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
		handleSymbols(ctx)
		time.Sleep(time.Second * 3)
	}
}

func handleSymbols(ctx model.Context) {
	mexcAPIData := mexc.Symbols(ctx)
	log.Infof("symbols count %d", len(mexcAPIData.Payload.(model.SupportSymbols).Symbols))
	for _, symbol := range mexcAPIData.Payload.(model.SupportSymbols).Symbols {
		symbolList := strings.Split(symbol, "_")
		exchangeSymbol := dao.ExchangeSymbol{ExchangeName: "mexc", Symbol: symbol, Symbol1: symbolList[0], Symbol2: symbolList[1]}

		err := redis.RdbClient(ctx).Set(context.Background(), symbol, symbol, 0).Err()
		if err != nil {
			panic(err)
		}

		mysql.MySQLClient(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&exchangeSymbol)
	}
}
