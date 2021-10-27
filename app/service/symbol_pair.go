package service

import (
	cinfra "cc-robot/core/tool/infra"
	"cc-robot/core/tool/mysql"
	"cc-robot/core/tool/redis"
	"cc-robot/dao"
	mexc "cc-robot/extern"
	"cc-robot/model"
	"context"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"strings"
)

var mexcSupportSymbolPair model.SupportSymbolPair

func HandleMexcSymbolPair(ctx model.Context) {
	mexcAPIData := mexc.SymbolPair(ctx)
	supportSymbolPair := mexcAPIData.Payload.(model.SupportSymbolPair)

	symbolPairMap := make(map[string][]string, len(supportSymbolPair.SymbolPairList))
	for _, symbolPair := range supportSymbolPair.SymbolPairList {
		symbol1And2 := strings.Split(symbolPair, "_")
		symbolPairMap[symbolPair] = symbol1And2
	}
	supportSymbolPair.SymbolPairMap = symbolPairMap

	oldSymbolPairCount := len(mexcSupportSymbolPair.SymbolPairList)
	newSymbolPairCount := len(supportSymbolPair.SymbolPairList)
	if oldSymbolPairCount > 0 && newSymbolPairCount > oldSymbolPairCount {
		handleSymbolPairAppear(ctx, supportSymbolPair.Exchange, mexcSupportSymbolPair, supportSymbolPair)
	}
	log.WithFields(log.Fields{
		"oldSymbolPairCount": oldSymbolPairCount,
		"newSymbolPairCount": newSymbolPairCount,
	}).Info("handleSymbolPair")

	for symbolPair, symbol1And2 := range supportSymbolPair.SymbolPairMap {
		exchangeSymbolPair := dao.ExchangeSymbolPair{ExchangeName: supportSymbolPair.Exchange, SymbolPair: symbolPair, Symbol1: symbol1And2[0], Symbol2: symbol1And2[1]}

		err := redis.RdbClient(ctx).Set(context.Background(), symbolPair, symbolPair, 0).Err()
		if err != nil {
			panic(err)
		}

		mysql.MySQLClient(ctx).Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&exchangeSymbolPair)
	}

	mexcSupportSymbolPair = supportSymbolPair
}

func handleSymbolPairAppear(ctx model.Context, exchange string, oldSupportSymbolPair model.SupportSymbolPair, newSupportSymbolPair model.SupportSymbolPair) {
	for symbolPair := range newSupportSymbolPair.SymbolPairMap {
		if _, ok := oldSupportSymbolPair.SymbolPairMap[symbolPair]; !ok {
			cinfra.GiantEventText(ctx, fmt.Sprintf("%s symbol pair appear %s", exchange, symbolPair))
		}
	}
}
