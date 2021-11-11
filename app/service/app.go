package service

import (
	cid "cc-robot/core/tool/id"
	clog "cc-robot/core/tool/log"
	cruntime "cc-robot/core/tool/runtime"
	"cc-robot/model"
	"context"
	"go.uber.org/zap"
	"math/big"
	"time"
)

type App struct {
	AppearSymbolPairCh      chan model.AppearSymbolPair
	AcquireBetterPriceCh    chan model.AppearSymbolPair
	processOrderManagerCh   chan model.SymbolPairBetterPrice
	AppearSymbolPairManager map[string]model.SymbolPairBetterPrice
	ListeningSymbolPair     map[string][]string
	adjustOrderFailed       map[string]bool
	SymbolPairConf          map[string]model.SymbolPairConf
}

func RunApp() *App {
	clog.EventLog.Info("run app")
	app := initApp()
	go app.initLogic()
	return app
}

func initApp() *App {
	app := &App{
		AppearSymbolPairCh:      make(chan model.AppearSymbolPair),
		AcquireBetterPriceCh:    make(chan model.AppearSymbolPair),
		processOrderManagerCh:   make(chan model.SymbolPairBetterPrice),
		AppearSymbolPairManager: map[string]model.SymbolPairBetterPrice{},
		ListeningSymbolPair:     make(map[string][]string),
		adjustOrderFailed:       make(map[string]bool),
		SymbolPairConf:          make(map[string]model.SymbolPairConf),
	}
	return app
}

func (app *App) initLogic() {
	go app.FetchSupportSymbolPairs()
	go app.GetAppearSymbolPairs()

	go app.listenBetterPrice()
	go app.listenOrderManager()
}

func (app *App) listenBetterPrice() {
	for {
		select {
		case appearSymbolPair := <-app.AcquireBetterPriceCh:
			ctx := clog.NewContext(context.TODO(),
				zap.String("symbolPair", appearSymbolPair.SymbolPair),
				zap.String(cruntime.FuncName() + "-traceId", cid.UniuqeId()),
			)

			if _, ok := app.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
				app.ListeningSymbolPair[appearSymbolPair.SymbolPair] = appearSymbolPair.Symbol1And2
				app.SymbolPairConf[appearSymbolPair.SymbolPair] = model.SymbolPairConf{
					// default cost is 10 USDT
					BidCost:            big.NewFloat(10),
					ExpectedProfitRate: big.NewFloat(0.1),
				}
				go app.ProcessMexcSymbolPairTicker(ctx, appearSymbolPair)
			} else {
				clog.WithCtxEventLog(ctx).With(zap.Reflect("appearSymbolPair", appearSymbolPair)).Error("listen better price exist")
			}
		}
	}
}

func (app *App) listenOrderManager() {
	for {
		select {
		case symbolPairBetterPrice := <-app.processOrderManagerCh:
			go app.ProcessOrder(symbolPairBetterPrice)
		}
	}
}

func (app *App) FetchSupportSymbolPairs() {
	for {
		fetchSupportSymbolPairs(*app, context.TODO())
		time.Sleep(10 * time.Minute)
	}
}

func (app *App) GetAppearSymbolPairs() {
	for {
		getAppearSymbolPairs(*app, context.TODO())
		time.Sleep(10 * time.Minute)
	}
}

func (app *App) ProcessMexcSymbolPairTicker(ctx context.Context, appearSymbolPair model.AppearSymbolPair) {
	if !app.shouldContinueBySupportSymbolPair(ctx, appearSymbolPair.Symbol1And2) {
		return
	}

	for {
		processMexcSymbolPairTicker(*app, ctx, appearSymbolPair)
	}
}

func (app *App) ProcessOrder(symbolPairBetterPrice model.SymbolPairBetterPrice) {
	if !app.shouldContinueBySupportSymbolPair(symbolPairBetterPrice.Ctx, symbolPairBetterPrice.AppearSymbolPair.Symbol1And2) {
		return
	}

	processMexcOrder(*app, symbolPairBetterPrice.Ctx, symbolPairBetterPrice)
}

func (app *App) shouldContinueBySupportSymbolPair(ctx context.Context, symbol1And2 []string) bool {
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
