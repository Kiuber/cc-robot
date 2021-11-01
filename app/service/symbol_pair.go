package service

import (
	cid "cc-robot/core/tool/id"
	"cc-robot/core/tool/mysql"
	"cc-robot/core/tool/redis"
	"cc-robot/dao"
	mexc "cc-robot/extern"
	"cc-robot/model"
	"context"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"math/big"
	"strconv"
	"strings"
)

var mexcSupportSymbolPair model.SupportSymbolPair

func processMexcAppearSymbolPair(app App) {
	mexcAPIData := mexc.SupportSymbolPair()
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
		findNewSymbolPairs(app, supportSymbolPair.Exchange, mexcSupportSymbolPair, supportSymbolPair)
	}
	log.WithFields(log.Fields{
		"oldSymbolPairCount": oldSymbolPairCount,
		"newSymbolPairCount": newSymbolPairCount,
	}).Info("new old symbol pair count")

	for symbolPair, symbol1And2 := range supportSymbolPair.SymbolPairMap {
		exchangeSymbolPair := dao.ExchangeSymbolPair{ExchangeName: supportSymbolPair.Exchange, SymbolPair: symbolPair, Symbol1: symbol1And2[0], Symbol2: symbol1And2[1]}

		err := redis.RdbClient().Set(context.Background(), symbolPair, symbolPair, 0).Err()
		if err != nil {
			panic(err)
		}

		mysql.MySQLClient().Clauses(clause.OnConflict{
			UpdateAll: true,
		}).Create(&exchangeSymbolPair)
	}

	mexcSupportSymbolPair = supportSymbolPair
}

func findNewSymbolPairs(app App, exchange string, oldSupportSymbolPair model.SupportSymbolPair, newSupportSymbolPair model.SupportSymbolPair) {
	for symbolPair, symbol1And2 := range newSupportSymbolPair.SymbolPairMap {
		if _, ok := oldSupportSymbolPair.SymbolPairMap[symbolPair]; !ok {
			appearSymbolPair := model.AppearSymbolPair{SymbolPair: symbolPair, Symbol1And2: symbol1And2, Exchange: exchange}
			log.WithFields(log.Fields{"appearSymbolPair": appearSymbolPair}).Info("appear symbol pair")
			app.symbolPairCh <- appearSymbolPair
		}
	}
}

func processMexcSymbolPairTicker(app App, appearSymbolPair model.AppearSymbolPair) {
	mexcAPIData := mexc.DepthInfo(appearSymbolPair.SymbolPair, "5")
	if !mexcAPIData.OK {
		return
	}

	depthInfo := mexcAPIData.Payload.(model.DepthInfo)

	asks := depthInfo.Asks
	lowestOfAsk := asks[0]
	float, err := strconv.ParseFloat(lowestOfAsk.Price, 64)
	if err != nil {
		return
	}

	lowestOfAskPrice := big.NewFloat(float)
	defaultBidUSDT := big.NewFloat(6)
	quantity := big.NewFloat(0)

	testBidPrice := big.NewFloat(0)
	testBidPrice.Quo(lowestOfAskPrice, big.NewFloat(2))

	quantity.Quo(defaultBidUSDT, testBidPrice)

	log.WithFields(log.Fields{
		"symbol":         appearSymbolPair.SymbolPair,
		"lowest_price":   lowestOfAskPrice,
		"test_ask_price": testBidPrice,
	}).Info("symbol lowest price")

	mexc.CreateOrder(model.Order{
		SymbolPair:    appearSymbolPair.SymbolPair,
		Price:         testBidPrice.String(),
		Quantity:      quantity.String(),
		TradeType:     "BID",
		OrderType:     "LIMIT_ORDER",
		ClientOrderId: cid.UniuqeId(),
	})
}
