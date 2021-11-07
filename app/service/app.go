package service

import (
	cinfra "cc-robot/core/tool/infra"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"fmt"
	"github.com/sirupsen/logrus"
	"math/big"
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
	clog.EventLog().Info("run app")
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
	go app.ProcessMexcAppearSymbolPair()
	go app.listenAppearSymbolPair()
	go app.listenBetterPrice()
	go app.listenOrderManager()
}

func (app *App) listenAppearSymbolPair() {
	for {
		select {
		case appearSymbolPair := <-app.AppearSymbolPairCh:
			cinfra.GiantEventText(fmt.Sprintf("%s appear %s symbol pair", appearSymbolPair.Exchange, appearSymbolPair.SymbolPair))
			app.AcquireBetterPriceCh <- appearSymbolPair
		}
	}
}

func (app *App) listenBetterPrice() {
	for {
		select {
		case appearSymbolPair := <-app.AcquireBetterPriceCh:
			if _, ok := app.ListeningSymbolPair[appearSymbolPair.SymbolPair]; !ok {
				app.ListeningSymbolPair[appearSymbolPair.SymbolPair] = appearSymbolPair.Symbol1And2
				app.SymbolPairConf[appearSymbolPair.SymbolPair] = model.SymbolPairConf{
					// default cost is 10 USDT
					BidCost:            big.NewFloat(10),
					ExpectedProfitRate: big.NewFloat(0.1),
				}
				go app.ProcessMexcSymbolPairTicker(appearSymbolPair)
			} else {
				clog.EventLog().WithFields(logrus.Fields{"appearSymbolPair": appearSymbolPair}).Error("listen better price exist")
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

func (app *App) ProcessMexcAppearSymbolPair() {
	for {
		processMexcAppearSymbolPair(*app)
	}
}

func (app *App) ProcessMexcSymbolPairTicker(appearSymbolPair model.AppearSymbolPair) {
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
		clog.EventLog().WithFields(logrus.Fields{
			"symbol1And2": symbol1And2,
		}).Error("not support symbol pair")
		return false
	}

	leftSymbol := symbol1And2[0]
	rightSymbol := symbol1And2[1]
	supportRightSymbol := "USDT"
	ok := rightSymbol == supportRightSymbol
	if !ok {
		clog.EventLog().WithFields(logrus.Fields{
			"leftSymbol":  leftSymbol,
			"rightSymbol": rightSymbol,
		}).Error("not support symbol pair")
	}
	return ok
}
