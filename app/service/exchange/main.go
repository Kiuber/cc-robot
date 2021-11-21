package s_exchange

import (
	cid "cc-robot/core/tool/id"
	clog "cc-robot/core/tool/log"
	"cc-robot/core/tool/mysql"
	"cc-robot/core/tool/redis"
	cruntime "cc-robot/core/tool/runtime"
	"cc-robot/dao"
	emexc "cc-robot/extern"
	"cc-robot/model"
	"context"
	"go.uber.org/zap"
	"gorm.io/gorm/clause"
	"strings"
)

type Exchange struct {
}

const (
	ExchangeNameMexc string = "mexc"
)

var ExchangeNames = []string{
	ExchangeNameMexc,
}

func Main() *Exchange {
	exchange := &Exchange{}
	return exchange
}

func (exchange *Exchange) SaveAPISupportSymbolPairsOfAllExchanges() {
	ctx := clog.NewContext(context.TODO(), zap.String(cruntime.FuncName()+"-traceId", cid.UniuqeId()))
	for _, exchangeName := range ExchangeNames {
		exchange.fetchAndUpsertAPISupportSymbolPairs(ctx, exchangeName)
	}
}

func (exchange *Exchange) fetchAndUpsertAPISupportSymbolPairs(ctx context.Context, exchangeName string) {
	mexcAPIData := emexc.SupportSymbolPair(ctx)
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

	clog.WithCtxEventLog(ctx).With(
		zap.Int("supportSymbolPair count", len(supportSymbolPair.SymbolPairList)),
	).Info("fetched symbol pairs")

	exchange.upsertSymbolPairs(exchangeName, supportSymbolPair)
}

func (exchange *Exchange) upsertSymbolPairs(exchangeName string, supportSymbolPair model.SupportSymbolPair) {
	for symbolPair, symbol1And2 := range supportSymbolPair.SymbolPairMap {
		exchangeSymbolPair := dao.ExchangeSymbolPair{ExchangeName: exchangeName, SymbolPair: symbolPair, Symbol1: symbol1And2[0], Symbol2: symbol1And2[1]}

		err := redis.RdbClient().Set(context.Background(), symbolPair, symbolPair, 0).Err()
		if err != nil {
			panic(err)
		}

		mysql.MySQLClient().Clauses(clause.OnConflict{
			DoNothing: true,
		}).Create(&exchangeSymbolPair)
	}
}
