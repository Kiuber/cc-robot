package service

import (
	cinfra "cc-robot/core/tool/infra"
	"cc-robot/model"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type App struct {
	appearSymbolPairManager map[string]model.SymbolPairBetterPrice
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
		appearSymbolPairManager: map[string]model.SymbolPairBetterPrice{},
		symbolPairCh:   make(chan model.AppearSymbolPair),
		BetterPriceCh:  make(chan model.AppearSymbolPair),
		orderManagerCh: make(chan model.SymbolPairBetterPrice),
	}
	return app
}

func(app *App) initLogic() {
	go app.ProcessMexcAppearSymbolPair()
	go app.listenAppearSymbolPair()
	go app.listenBetterPrice()
	go app.listenOrderManager()
}

func(app *App) listenAppearSymbolPair() {
	for {
		select {
		case appearSymbolPair := <-app.symbolPairCh:
			cinfra.GiantEventText(fmt.Sprintf("%s appear %s symbol pair", appearSymbolPair.Exchange, appearSymbolPair.SymbolPair))
			app.BetterPriceCh <- appearSymbolPair
		}
	}
}

func(app *App) listenBetterPrice() {
	for {
		select {
		case appearSymbolPair := <-app.BetterPriceCh:
			if _, ok := app.appearSymbolPairManager[appearSymbolPair.SymbolPair]; !ok {
				go app.ProcessMexcSymbolPairTicker(appearSymbolPair)
			} else {
				log.WithFields(log.Fields{"appearSymbolPair": appearSymbolPair}).Error("listen better price exist")
			}
		}
	}
}

func(app *App) listenOrderManager() {
	for {
		select {
		case symbolPairBetterPrice := <-app.orderManagerCh:
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
	if !app.shouldContinueBySupportSymbolPair(appearSymbolPair.Symbol1And2) {
		return
	}

	for {
		processMexcSymbolPairTicker(*app, appearSymbolPair)
	}
}

func (app *App) ProcessOrder(symbolPairBetterPrice model.SymbolPairBetterPrice) {
	if !app.shouldContinueBySupportSymbolPair(symbolPairBetterPrice.AppearSymbolPair.Symbol1And2) {
		return
	}

	processMexcOrder(*app, symbolPairBetterPrice)
}

func (app *App) shouldContinueBySupportSymbolPair(symbol1And2 []string) bool {
	if len(symbol1And2) != 2 {
		log.WithFields(log.Fields{
			"symbol1And2": symbol1And2,
		}).Error("not support symbol pair")
		return false
	}

	leftSymbol := symbol1And2[0]
	rightSymbol := symbol1And2[1]
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
