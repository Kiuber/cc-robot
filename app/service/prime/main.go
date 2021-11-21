package s_prime

import (
	cboot "cc-robot/core/boot"
	cid "cc-robot/core/tool/id"
	cinfra "cc-robot/core/tool/infra"
	clog "cc-robot/core/tool/log"
	"cc-robot/core/tool/mysql"
	cruntime "cc-robot/core/tool/runtime"
	"cc-robot/dao"
	emexc "cc-robot/extern"
	"cc-robot/model"
	s_exchange "cc-robot/service/exchange"
	"context"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"math/big"
	"strconv"
	"strings"
	"time"
)

type Prime struct {
	AppearSymbolPairCh      chan model.AppearSymbolPair
	AcquireBetterPriceCh    chan model.AppearSymbolPair
	ProcessOrderManagerCh   chan model.SymbolPairBetterPrice
	AppearSymbolPairManager map[string]model.SymbolPairBetterPrice
	ListeningSymbolPair     map[string][]string
	AdjustOrderFailed       map[string]bool
	SymbolPairConf          map[string]model.SymbolPairConf
}

func (prime *Prime) updateSymbolPairConf(symbolPair string, symbolPairConf model.SymbolPairConf) {
	prime.SymbolPairConf[symbolPair] = symbolPairConf
}

func Main() *Prime {
	clog.EventLog.Info("run prime")
	prime := initPrime()
	go prime.initLogic()
	return prime
}

func initPrime() *Prime {
	prime := &Prime{
		AppearSymbolPairCh:      make(chan model.AppearSymbolPair),
		AcquireBetterPriceCh:    make(chan model.AppearSymbolPair),
		ProcessOrderManagerCh:   make(chan model.SymbolPairBetterPrice),
		AppearSymbolPairManager: map[string]model.SymbolPairBetterPrice{},
		ListeningSymbolPair:     make(map[string][]string),
		AdjustOrderFailed:       make(map[string]bool),
		SymbolPairConf:          make(map[string]model.SymbolPairConf),
	}
	return prime
}

func (prime *Prime) initLogic() {
	go prime.listenBetterPrice()
	go prime.listenBuyOrTakeProfit()
}

func (prime *Prime) listenBetterPrice() {
	for {
		select {
		case appearSymbolPair := <-prime.AcquireBetterPriceCh:
			ctx := clog.NewContext(context.TODO(),
				zap.String("symbolPair", appearSymbolPair.SymbolPair),
				zap.String(cruntime.FuncName()+"-traceId", cid.UniuqeId()),
			)

			if _, ok := prime.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
				prime.ListeningSymbolPair[appearSymbolPair.SymbolPair] = appearSymbolPair.Symbol1And2
				prime.SymbolPairConf[appearSymbolPair.SymbolPair] = model.SymbolPairConf{
					// default cost is 10 USDT
					TotalDealCost:      big.NewFloat(0),
					BidDiffCost:        big.NewFloat(0),
					BidCost:            big.NewFloat(10),
					ExpectedProfitRate: big.NewFloat(0.1),
				}
				go prime.getBetterPriceLoop(ctx, appearSymbolPair)
			} else {
				clog.WithCtxEventLog(ctx).With(zap.Reflect("appearSymbolPair", appearSymbolPair)).Error("listen better price exist")
			}
		}
	}
}

func (prime *Prime) listenBuyOrTakeProfit() {
	for {
		select {
		case symbolPairBetterPrice := <-prime.ProcessOrderManagerCh:
			prime.buyOrTakeProfit(symbolPairBetterPrice)
		}
	}
}

func (prime *Prime) TryUpdatePrimeSymbolPair() {
	ctx := clog.NewContext(context.TODO(), zap.String(cruntime.FuncName()+"-traceId", cid.UniuqeId()))

	var exchangePrimeConfig []dao.ExchangePrimeConfig
	mysql.MySQLClient().Where("status = 'enabled'").Find(&exchangePrimeConfig)

	logger := clog.WithCtxEventLog(ctx)
	logger.Info("fetch prime config")

	for symbolPair := range prime.ListeningSymbolPair {
		symbolPairDisabled := false
		for _, primeConfig := range exchangePrimeConfig {
			if primeConfig.SymbolPair == symbolPair {
				symbolPairDisabled = true
				break
			}
		}

		if !symbolPairDisabled {
			logger.With(zap.String("symbolPair", symbolPair)).Info("remove symbol pair for prime")
			delete(prime.ListeningSymbolPair, symbolPair)
			delete(prime.AppearSymbolPairManager, symbolPair)
			delete(prime.SymbolPairConf, symbolPair)
		}
	}

	for _, primeConfig := range exchangePrimeConfig {
		symbol1And2 := strings.Split(primeConfig.SymbolPair, "_")

		if _, ok := prime.ListeningSymbolPair[primeConfig.SymbolPair]; ok {
			continue
		}

		if !prime.isSupportThisSymbolPair(ctx, symbol1And2) {
			continue
		}

		appearSymbolPair := model.AppearSymbolPair{SymbolPair: primeConfig.SymbolPair, Symbol1And2: symbol1And2}

		prime.AcquireBetterPriceCh <- appearSymbolPair
		prime.AppearSymbolPairManager[primeConfig.SymbolPair] = model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair}
	}
	logger.With(zap.Reflect("symbolPairConf", prime.SymbolPairConf)).Info("fetch prime config finished")
}

func (prime *Prime) getBetterPriceLoop(ctx context.Context, appearSymbolPair model.AppearSymbolPair) {
	if !prime.isSupportThisSymbolPair(ctx, appearSymbolPair.Symbol1And2) {
		return
	}

	for {
		if _, ok := prime.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
			clog.WithCtxEventLog(ctx).With(zap.String("symbolPair", appearSymbolPair.SymbolPair)).Info("exit get better price loop")
			break
		}
		prime.getBetterPrice(ctx, appearSymbolPair)
	}
}

func (prime *Prime) buyOrTakeProfit(symbolPairBetterPrice model.SymbolPairBetterPrice) {
	if !prime.isSupportThisSymbolPair(symbolPairBetterPrice.Ctx, symbolPairBetterPrice.AppearSymbolPair.Symbol1And2) {
		return
	}

	prime.doBuyOrTakeProfit(symbolPairBetterPrice.Ctx, symbolPairBetterPrice)
}

func (prime *Prime) CheckAndAlarmSymbolPairsOfAllExchanges() {
	for _, exchangeName := range s_exchange.ExchangeNames {
		prime.checkAndAlarmForSymbolPairs(exchangeName)
	}
}

func (prime *Prime) checkAndAlarmForSymbolPairs(exchangeName string) {
	ctx := clog.NewContext(context.TODO(), zap.String(cruntime.FuncName()+"-traceId", cid.UniuqeId()))

	var exchangeSymbolPairList []dao.ExchangeSymbolPair
	mysql.MySQLClient().Where("open_timestamp = 0").Find(&exchangeSymbolPairList)
	clog.WithCtxEventLog(ctx).With(zap.Int("exchangeSymbolPairList count", len(exchangeSymbolPairList))).Info("not open symbol pair yet")

	for _, pair := range exchangeSymbolPairList {
		logger := clog.WithCtxEventLog(ctx).With(zap.String("symbolPair", pair.SymbolPair))

		mexcAPIData := emexc.SymbolPairInfo(ctx, pair.SymbolPair)
		if !mexcAPIData.OK {
			logger.With(zap.String("err", mexcAPIData.Msg)).Debug("get symbol pair info failed")
			continue
		}

		symbolPairInfo := mexcAPIData.Payload.(model.SymbolPairInfo)
		symbolPairInfo.WebLink = "https://www.mexc.com/zh-CN/exchange"
		if len(symbolPairInfo.OpenTime) == 0 {
			limit := int64(5)
			mexcAPIData = emexc.KLine(ctx, pair.SymbolPair, "1m", strconv.FormatInt(time.Now().Unix()-((limit+1)*60), 10), strconv.FormatInt(limit, 10))
			if mexcAPIData.OK {
				kLineData := mexcAPIData.Payload.([]interface{})
				if len(kLineData) > 0 {
					logger.Debug("has kline data but no openTime")
				}
			} else {
				logger.With(zap.String("msg", mexcAPIData.Msg)).Debug("get kline failed")
			}
			continue
		}

		openTime, err := time.ParseInLocation(time.RFC3339, symbolPairInfo.OpenTime, time.Local)
		if err != nil {
			logger.With(zap.String("err", err.Error())).Error("parse time failed")
			continue
		}

		var exchangeSymbolPair dao.ExchangeSymbolPair
		exchangeSymbolPair.ExchangeName = pair.ExchangeName
		exchangeSymbolPair.OpenTimestamp = int(openTime.Unix())
		if openTime.Unix() > time.Now().Unix() {
			clog.WithCtxEventLog(ctx).With(zap.String("symbolPair", pair.SymbolPair)).Info("new symbol pair")
			msg := fmt.Sprintf("%s 出现 %s 交易对，链接：%s/%s, 开盘时间: %s", pair.ExchangeName, pair.SymbolPair, symbolPairInfo.WebLink, pair.SymbolPair, openTime.Format(time.RFC3339Nano))
			cinfra.GiantEventText(ctx, msg)
		}

		mysql.MySQLClient().Where("symbol_pair = ?", pair.SymbolPair).Updates(exchangeSymbolPair)
	}
}

func (prime *Prime) getBetterPrice(ctx context.Context, appearSymbolPair model.AppearSymbolPair) {
	mexcAPIData := emexc.DepthInfo(ctx, appearSymbolPair.SymbolPair, "5")
	if !mexcAPIData.OK {
		return
	}

	depthInfo := mexcAPIData.Payload.(model.DepthInfo)

	asks := depthInfo.Asks
	if asks == nil {
		clog.WithCtxEventLog(ctx).Error("asks is nil")
		return
	}
	if len(asks) <= 0 {
		clog.WithCtxEventLog(ctx).Error("asks is empty")
		return
	}

	lowestOfAsk := asks[0]
	float, err := strconv.ParseFloat(lowestOfAsk.Price, 64)
	if err != nil {
		clog.WithCtxEventLog(ctx).Error("parse float failed")
		return
	}
	lowestOfAskPrice := big.NewFloat(float)

	oldLowestOfAskPrice := prime.AppearSymbolPairManager[appearSymbolPair.SymbolPair].LowestOfAskPrice

	if _, ok := prime.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
		return
	}

	if oldLowestOfAskPrice == nil || lowestOfAskPrice.Cmp(oldLowestOfAskPrice) != 0 {
		prime.ProcessOrderManagerCh <- model.SymbolPairBetterPrice{
			AppearSymbolPair: appearSymbolPair, LowestOfAskPrice: lowestOfAskPrice,
			Ctx: clog.NewContext(ctx, zap.String(cruntime.FuncName()+"-traceId", cid.UniuqeId())),
		}
		prime.AppearSymbolPairManager[appearSymbolPair.SymbolPair] = model.SymbolPairBetterPrice{AppearSymbolPair: appearSymbolPair, LowestOfAskPrice: lowestOfAskPrice}
		clog.WithCtxEventLog(ctx).With(
			zap.Reflect("old price", prime.AppearSymbolPairManager[appearSymbolPair.SymbolPair].LowestOfAskPrice),
			zap.Reflect("new price", lowestOfAskPrice),
		).Info("better price need update")
	}
}

func (prime *Prime) doBuyOrTakeProfit(ctx context.Context, symbolPairBetterPrice model.SymbolPairBetterPrice) {
	logger := clog.WithCtxEventLog(ctx)
	logger.Info("prepare doBuyOrTakeProfit")

	symbolPair := symbolPairBetterPrice.AppearSymbolPair.SymbolPair
	symbolPairConf := prime.SymbolPairConf[symbolPair]

	bidOrderList, err := prime.getOrderList(ctx, symbolPair, "BID")
	if err != nil {
		logger.Error(err.Error())
		return
	}

	bidDiffCost := big.NewFloat(0)
	bidDiffCost = bidDiffCost.Add(bidDiffCost, symbolPairConf.BidCost)
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

	symbolPairConf.UpdateTotalDealCost(totalDealCost)
	lowestOfAskPrice := symbolPairBetterPrice.LowestOfAskPrice

	// cancel all orders of the symbol pair
	mexcAPIData := emexc.CancelOrder(ctx, symbolPair)
	if !mexcAPIData.OK {
		logger.Error("cancel order failed")
		prime.AdjustOrderFailed[symbolPair] = false
		return
	} else {
		prime.AdjustOrderFailed[symbolPair] = true
	}

	testBidPrice := lowestOfAskPrice
	if cboot.GV.IsDev {
		// testBidPrice = big.NewFloat(0)
		// testBidPrice.Quo(lowestOfAskPrice, big.NewFloat(2))
	}

	// bid finished: deal 90% cost
	totalDealCostRate.Quo(totalDealCost, bidDiffCost)
	if totalDealCostRate.Cmp(big.NewFloat(0.9)) < 0 {
		logger.Info("prepare try add position")

		bidDiffCost.Sub(bidDiffCost, totalDealCost)
		quantity := big.NewFloat(0)
		quantity.Quo(bidDiffCost, testBidPrice)

		symbolPairConf.BidDiffCost = bidDiffCost
		prime.updateSymbolPairConf(symbolPair, symbolPairConf)

		clog.WithCtxEventLog(ctx).With(
			zap.Any("prime", prime.SymbolPairConf),
			zap.Any("symbolPairConf", symbolPairConf),
			zap.Any("bidDiffCost", bidDiffCost),
			zap.String("symbol", symbolPair),
			zap.Any("quantity", quantity),
			zap.Any("lowestOfAskPrice", lowestOfAskPrice),
			zap.Any("testBidPrice", testBidPrice),
			zap.Any("totalDealCost", totalDealCost),
			zap.Any("totalDealCostRate", totalDealCostRate),
			zap.Any("totalHoldQuantity", totalHoldQuantity),
		).Info("prepare bid detail")

		mexcAPIData = prime.adjustPosition(ctx, symbolPair, "BID", testBidPrice, quantity)
		if mexcAPIData.OK {
			logger.Info("create order is ok")
			prime.AdjustOrderFailed[symbolPair] = true
		} else {
			logger.With(zap.String("err", mexcAPIData.Msg)).Error("create order is failed")
			prime.AdjustOrderFailed[symbolPair] = false
		}
	} else {
		logger.Info("prepare try sub position")

		if lowestOfAskPrice.Cmp(big.NewFloat(0)) <= 0 {
			logger.Error("lowest ask price is <= 0")
			return
		}
		mexcAPIData = emexc.AccountInfo(ctx)
		accountInfo := mexcAPIData.Payload.(model.AccountInfo)
		if _, ok := accountInfo[symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[0]]; !ok {
			logger.Info("not hold")
			return
		}

		balanceInfo := accountInfo[symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[0]]
		holdQuantityFloat, err := strconv.ParseFloat(balanceInfo.Available, 64)
		if err != nil {
			clog.WithCtxEventLog(ctx).With(zap.Reflect("account symbol pair info", balanceInfo)).Error("parse float failed")
			return
		}
		holdQuantity := big.NewFloat(holdQuantityFloat)

		if holdQuantity.Cmp(big.NewFloat(0)) <= 0 {
			clog.WithCtxEventLog(ctx).With(zap.Reflect("balanceInfo", balanceInfo)).Error("not hold")
			return
		}

		totalHoldCost := big.NewFloat(0)
		totalProfit := big.NewFloat(0)
		totalProfitRate := big.NewFloat(0)
		expectedProfitRate := symbolPairConf.ExpectedProfitRate
		withExpectedProfitRateDiff := big.NewFloat(0)
		totalHoldCost.Mul(testBidPrice, holdQuantity)
		totalProfit.Sub(totalHoldCost, totalDealCost)
		totalProfitRate.Quo(totalProfit, totalDealCost)

		withExpectedProfitRateDiff.Sub(totalProfitRate, expectedProfitRate)
		clog.WithCtxEventLog(ctx).With(
			zap.Reflect("holdQuantity", holdQuantity),
			zap.Reflect("totalDealCost", totalDealCost),
			zap.Reflect("totalHoldCost", totalHoldCost),
			zap.Reflect("totalProfit", totalProfit),
			zap.Reflect("totalProfitRate", totalProfitRate),
			zap.Reflect("expectedProfitRate", expectedProfitRate),
			zap.Reflect("withExpectedProfitRateDiff", withExpectedProfitRateDiff),
		).Info("profit detail")
		hasReachProfit := totalProfitRate.Cmp(expectedProfitRate) >= 0
		logger.With(zap.Bool("reached?", hasReachProfit)).Info("has reach expected profit rate")
		if hasReachProfit {
			mexcAPIData = prime.adjustPosition(ctx, symbolPair, "ASK", testBidPrice, holdQuantity)
			if mexcAPIData.OK {
				logger.Info("create order is ok")
				prime.AdjustOrderFailed[symbolPair] = true
			} else {
				logger.With(zap.String("err", mexcAPIData.Msg)).Error("create order is failed")
				prime.AdjustOrderFailed[symbolPair] = false
			}
		}
	}
}

func (prime *Prime) isSupportThisSymbolPair(ctx context.Context, symbol1And2 []string) bool {
	if len(symbol1And2) != 2 {
		clog.WithCtxEventLog(ctx).Error("not support symbol pair")
		return false
	}

	rightSymbol := symbol1And2[1]
	supportRightSymbol := "USDT"
	ok := rightSymbol == supportRightSymbol
	if !ok {
		clog.WithCtxEventLog(ctx).Error("not support symbol pair")
	}
	return ok
}

func (prime *Prime) adjustPosition(ctx context.Context, symbolPair string, tradeType string, price *big.Float, quantity *big.Float) model.MexcAPIData {
	mexcAPIData := emexc.CreateOrder(ctx, model.Order{
		SymbolPair:    symbolPair,
		Price:         price.String(),
		Quantity:      quantity.String(),
		TradeType:     tradeType,
		OrderType:     "LIMIT_ORDER",
		ClientOrderId: cid.UniuqeId(),
	})
	return mexcAPIData
}

func (prime *Prime) getOrderList(ctx context.Context, symbolPair string, tradeType string) (orderList model.OrderList, err error) {
	states := []string{"FILLED", "PARTIALLY_FILLED", "PARTIALLY_CANCELED"}
	for _, state := range states {
		mexcAPIData := emexc.OrderList(ctx, symbolPair, tradeType, state, "1000", "")
		if !mexcAPIData.OK {
			return nil, errors.New("order list inconsistency: " + mexcAPIData.Msg)
		}
		orderList = append(orderList, mexcAPIData.Payload.(model.OrderList)...)
	}
	return orderList, nil
}
