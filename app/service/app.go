package service

import (
	cinfra "cc-robot/core/tool/infra"
	"cc-robot/model"
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
)

type App struct {
	symbolPairCh   chan model.AppearSymbolPair
	BetterPriceCh  chan model.AppearSymbolPair
	orderManagerCh chan model.SymbolPairBetterPrice
}

func RunApp() *App {
	log.Info("run app")
	app := initApp()
	go app.initLogic()
	return app
}

func initApp() *App {
	app := &App{
		symbolPairCh:   make(chan model.AppearSymbolPair),
		BetterPriceCh:  make(chan model.AppearSymbolPair),
		orderManagerCh: make(chan model.SymbolPairBetterPrice),
	}
	return app
}

func(app *App) initLogic() {
	go app.ProcessMexcAppearSymbolPair()
	for {
		select {
		case appearSymbolPair := <-app.symbolPairCh:
			cinfra.GiantEventText(fmt.Sprintf("%s appear %s symbol pair", appearSymbolPair.Exchange, appearSymbolPair.SymbolPair))
			app.BetterPriceCh <- appearSymbolPair

		case appearSymbolPair := <- app.BetterPriceCh:
			go app.ProcessMexcSymbolPairTicker(appearSymbolPair)

		case symbolPairBetterPrice := <- app.orderManagerCh:
			go app.ProcessOrder(symbolPairBetterPrice)
		}
	}
}

func(app *App) ProcessMexcAppearSymbolPair() {
	for {
		processMexcAppearSymbolPair(*app)
	}
}

func(app *App) ProcessMexcSymbolPairTicker(appearSymbolPair model.AppearSymbolPair) {
	if !app.shouldContinueBySupportSymbolPair(appearSymbolPair.Symbol1And2[0], appearSymbolPair.Symbol1And2[1]) {
		return
	}

	for {
		processMexcSymbolPairTicker(*app, appearSymbolPair)
	}
}

func (app *App) ProcessOrder(symbolPairBetterPrice model.SymbolPairBetterPrice) {
	if !app.shouldContinueBySupportSymbolPair(symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[0], symbolPairBetterPrice.AppearSymbolPair.Symbol1And2[1]) {
		return
	}

	for {
		processMexcOrder(*app, symbolPairBetterPrice)
		time.Sleep(3 * time.Second)
	}
}

func (app *App) shouldContinueBySupportSymbolPair(leftSymbol string, rightSymbol string) bool {
	supportRightSymbol := "USDT"
	ok := rightSymbol == supportRightSymbol
	if !ok {
		log.WithFields(log.Fields{
			"leftSymbol": leftSymbol,
			"rightSymbol": rightSymbol,
		}).Error("not support symbol pair")
	}
	return ok
}
