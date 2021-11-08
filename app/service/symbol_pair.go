package service

import (
	cboot "cc-robot/core/boot"
	cid "cc-robot/core/tool/id"
	clog "cc-robot/core/tool/log"
	"cc-robot/core/tool/mysql"
	"cc-robot/core/tool/redis"
	"cc-robot/dao"
	mexc "cc-robot/extern"
	"cc-robot/model"
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"math/big"
	"strconv"
	"strings"
	"time"
)

var mexcSupportSymbolPair model.SupportSymbolPair

func processMexcAppearSymbolPair(app App) {
	mexcAPIData := mexc.SupportSymbolPair()
	if !mexcAPIData.OK {
		return
	}

	supportSymbolPair := mexcAPIData.Payload.(model.SupportSymbolPair)

	symbolPairMap := make(map[string][]string, len(supportSymbolPair.SymbolPairList))
	for _, symbolPair := range supportSymbolPair.SymbolPairList {
		symbol1And2 := strings.Split(symbolPair, "_")
		symbolPairMap[symbolPair] = symbol1And2
	}
	supportSymbolPair.SymbolPairMap = symbolPairMap

	oldSymbolPairCount := len(mexcSupportSymbolPair.SymbolPairList)
	newSymbolPairCount := len(supportSymbolPair.SymbolPairList)
	clog.EventLog().With(
		zap.Int("oldSymbolPairCount", oldSymbolPairCount),
		zap.Int("newSymbolPairCount", newSymbolPairCount),
	).Info("new and old symbol pair count")

	if newSymbolPairCount > oldSymbolPairCount {
		if oldSymbolPairCount > 0 {
			findNewSymbolPairs(app, supportSymbolPair.Exchange, mexcSupportSymbolPair, supportSymbolPair)
		}
		persistentSymbolPairs(supportSymbolPair)
	}

	mexcSupportSymbolPair = supportSymbolPair
}

func findNewSymbolPairs(app App, exchange string, oldSupportSymbolPair model.SupportSymbolPair, newSupportSymbolPair model.SupportSymbolPair) {
	for symbolPair, symbol1And2 := range newSupportSymbolPair.SymbolPairMap {
		if _, ok := oldSupportSymbolPair.SymbolPairMap[symbolPair]; !ok {
			limit := int64(5)
			mexcAPIData := mexc.KLine(symbolPair, "1m", strconv.FormatInt(time.Now().Unix()-((limit+1)*60), 10), strconv.FormatInt(limit, 10))
			if !mexcAPIData.OK {
				clog.EventLog().With(zap.String("msg", mexcAPIData.Msg)).Info("get kline failed")
				continue
			}

			kLineData := mexcAPIData.Payload.([]interface{})
			if len(kLineData) > 0 {
				clog.EventLog().With(zap.String("symbolPair", symbolPair)).Error("It doesn't look like a new symbolPair")
				continue
			}

			appearSymbolPair := model.AppearSymbolPair{SymbolPair: symbolPair, Symbol1And2: symbol1And2, Exchange: exchange}
			clog.EventLog().With(zap.Reflect("appearSymbolPair", appearSymbolPair)).Info("appear symbol pair")

			if _, ok := app.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
				app.AppearSymbolPairCh <- appearSymbolPair
				app.AppearSymbolPairManager[symbolPair] = model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair}
			}
		}
	}
}

func persistentSymbolPairs(supportSymbolPair model.SupportSymbolPair) {
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
}

func processMexcSymbolPairTicker(app App, appearSymbolPair model.AppearSymbolPair) {
	mexcAPIData := mexc.DepthInfo(appearSymbolPair.SymbolPair, "5")
	if !mexcAPIData.OK {
		return
	}

	depthInfo := mexcAPIData.Payload.(model.DepthInfo)

	asks := depthInfo.Asks
	if asks == nil {
		clog.EventLog().Error("asks is nil")
		return
	}
	if len(asks) <= 0 {
		clog.EventLog().Info("asks is empty")
		return
	}

	lowestOfAsk := asks[0]
	float, err := strconv.ParseFloat(lowestOfAsk.Price, 64)
	if err != nil {
		clog.EventLog().Error("parse float failed")
		return
	}
	lowestOfAskPrice := big.NewFloat(float)

	oldLowestOfAskPrice := app.AppearSymbolPairManager[appearSymbolPair.SymbolPair].LowestOfAskPrice

	logger := clog.EventLog().With(
		zap.String("symbolPair", appearSymbolPair.SymbolPair),
		zap.Reflect("old price", app.AppearSymbolPairManager[appearSymbolPair.SymbolPair].LowestOfAskPrice),
		zap.Reflect("new price", lowestOfAskPrice),
	)
	if oldLowestOfAskPrice == nil || lowestOfAskPrice.Cmp(oldLowestOfAskPrice) != 0 || !app.adjustOrderFailed[appearSymbolPair.SymbolPair] {
		// TODO: @qingbao, close previous app.processOrderManagerCh
		app.processOrderManagerCh <- model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair, LowestOfAskPrice: lowestOfAskPrice}
		app.AppearSymbolPairManager[appearSymbolPair.SymbolPair] = model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair, LowestOfAskPrice: lowestOfAskPrice}
		logger.Info("better price need update")
	} else {
		logger.Info("better price not need update")
	}
}

func processMexcOrder(app App, symbolPairBetterPrice model.SymbolPairBetterPrice) {
	symbolPair := symbolPairBetterPrice.AppearSymbolPair.SymbolPair
	symbolPairConf := app.SymbolPairConf[symbolPair]
	bidOrderList := getOrderList(symbolPair, "BID")
	logger := clog.EventLog().With(zap.String("symbolPair", symbolPair))

	bidCost := symbolPairConf.BidCost
	totalDealCost := big.NewFloat(0)
	totalDealCostRate := big.NewFloat(0)
	totalHoldQuantity := big.NewFloat(0)
	for _, order := range bidOrderList {
		dealCost, err := strconv.ParseFloat(order.DealCost, 64)
		if err != nil {
			logger.Error("parse float failed")
			return
		}
		dealQuantity, err := strconv.ParseFloat(order.DealQuantity, 64)
		if err != nil {
			logger.Error("parse float failed")
			return
		}
		totalDealCost.Add(totalDealCost, big.NewFloat(dealCost))
		totalHoldQuantity.Add(totalHoldQuantity, big.NewFloat(dealQuantity))
	}

	lowestOfAskPrice := symbolPairBetterPrice.LowestOfAskPrice

	// cancel all orders of the symbol pair
	mexcAPIData := mexc.CancelOrder(symbolPair)
	if !mexcAPIData.OK {
		logger.Error("cancel order failed")
		app.adjustOrderFailed[symbolPair] = false
		return
	} else {
		logger.Info("cancel order succeed")
		app.adjustOrderFailed[symbolPair] = true
	}

	testBidPrice := lowestOfAskPrice
	if cboot.GV.IsDev {
		// testBidPrice = big.NewFloat(0)
		// testBidPrice.Quo(lowestOfAskPrice, big.NewFloat(2))
	}

	// bid finished: deal 90% cost
	totalDealCostRate.Quo(totalDealCost, bidCost)
	if totalDealCostRate.Cmp(big.NewFloat(0.9)) < 0 {
		logger.Info("add position")

		bidCost.Sub(bidCost, totalDealCost)
		quantity := big.NewFloat(0)
		quantity.Quo(bidCost, testBidPrice)

		clog.EventLog().With(
			zap.Any("bidCost", bidCost),
			zap.String("symbol", symbolPair),
			zap.Any("quantity", quantity),
			zap.Any("lowestOfAskPrice", lowestOfAskPrice),
			zap.Any("testBidPrice", testBidPrice),
			zap.Any("totalDealCost", totalDealCost),
			zap.Any("totalDealCostRate", totalDealCostRate),
			zap.Any("totalHoldQuantity", totalHoldQuantity),
		).Info("prepare bid detail")

		mexcAPIData = adjustPosition(symbolPair, "BID", testBidPrice, quantity)
		if mexcAPIData.OK {
			logger.Info("create order is ok")
			app.adjustOrderFailed[symbolPair] = true
		} else {
			logger.Error("create order is failed")
			app.adjustOrderFailed[symbolPair] = false
		}
	} else {
		logger.Info("sub position")

		if lowestOfAskPrice.Cmp(big.NewFloat(0)) <= 0 {
			logger.Error("lowest ask price is <= 0")
			return
		}
		mexcAPIData = mexc.AccountInfo()
		accountInfo := mexcAPIData.Payload.(model.AccountInfo)
		if _, ok := accountInfo[symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[0]]; !ok {
			logger.Info("not hold")
			return
		}

		balanceInfo := accountInfo[symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[0]]
		holdQuantityFloat, err := strconv.ParseFloat(balanceInfo.Available, 64)
		if err != nil {
			clog.EventLog().With(zap.Reflect("account symbol pair info", balanceInfo)).Error("parse float failed")
			return
		}
		holdQuantity := big.NewFloat(holdQuantityFloat)

		if holdQuantity.Cmp(big.NewFloat(0)) <= 0 {
			clog.EventLog().With(zap.Reflect("balanceInfo", balanceInfo)).Error("not hold")
			return
		}

		totalHoldCost := big.NewFloat(0)
		totalProfit := big.NewFloat(0)
		totalProfitRate := big.NewFloat(0)
		expectedProfitRate := symbolPairConf.ExpectedProfitRate
		profitRateDiff := big.NewFloat(0)
		totalHoldCost.Mul(testBidPrice, holdQuantity)
		totalProfit.Sub(totalHoldCost, totalDealCost)
		totalProfitRate.Quo(totalProfit, totalDealCost)

		profitRateDiff.Sub(totalProfitRate, expectedProfitRate)
		clog.EventLog().With(
			zap.Reflect("holdQuantity", holdQuantity),
			zap.Reflect("totalDealCost", totalDealCost),
			zap.Reflect("totalHoldCost", totalHoldCost),
			zap.Reflect("totalProfit", totalProfit),
			zap.Reflect("totalProfitRate", totalProfitRate),
			zap.Reflect("expectedProfitRate", expectedProfitRate),
			zap.Reflect("profitRateDiff", profitRateDiff),
		).Info("profit detail")
		hasReachProfit := totalProfitRate.Cmp(expectedProfitRate) >= 0
		logger.With(zap.Bool("reached?", hasReachProfit)).Info("has reach expected profit rate")
		if hasReachProfit {
			mexcAPIData = adjustPosition(symbolPair, "ASK", testBidPrice, holdQuantity)
			if mexcAPIData.OK {
				logger.Info("create order is ok")
				app.adjustOrderFailed[symbolPair] = true
			} else {
				logger.Error("create order is failed")
				app.adjustOrderFailed[symbolPair] = false
			}
		} else {
			logger.Info("not reach expected profit rate")
		}
	}
}

func adjustPosition(symbolPair string, tradeType string, price *big.Float, quantity *big.Float) model.MexcAPIData {
	mexcAPIData := mexc.CreateOrder(model.Order{
		SymbolPair:    symbolPair,
		Price:         price.String(),
		Quantity:      quantity.String(),
		TradeType:     tradeType,
		OrderType:     "LIMIT_ORDER",
		ClientOrderId: cid.UniuqeId(),
	})
	return mexcAPIData
}

func getOrderList(symbolPair string, tradeType string) model.OrderList {
	var orderList model.OrderList
	states := []string{"FILLED", "PARTIALLY_FILLED", "PARTIALLY_CANCELED"}
	for _, state := range states {
		mexcAPIData := mexc.OrderList(symbolPair, tradeType, state, "1000", "")
		orderList = append(orderList, mexcAPIData.Payload.(model.OrderList)...)
	}
	return orderList
}
