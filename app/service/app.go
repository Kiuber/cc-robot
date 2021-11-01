package service

import (
	cinfra "cc-robot/core/tool/infra"
	"cc-robot/model"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type App struct {
	symbolPairCh chan model.AppearSymbolPair
	betterPriceCh chan model.AppearSymbolPair
}

func RunApp() {
	log.Info("run app")
	app := initApp()
	app.initLogic()
}

func initApp() *App {
	app := &App{
		symbolPairCh: make(chan model.AppearSymbolPair),
		betterPriceCh: make(chan model.AppearSymbolPair),
	}
	return app
}

func(app *App) initLogic() {
	go app.ProcessMexcAppearSymbolPair()
	for {
		select {
		case appearSymbolPair := <-app.symbolPairCh:
			app.betterPriceCh <- appearSymbolPair
			cinfra.GiantEventText(fmt.Sprintf("%s appear %s symbol pair", appearSymbolPair.Exchange, appearSymbolPair.SymbolPair))
		case appearSymbolPair := <- app.betterPriceCh:
			go app.ProcessMexcSymbolPairTicker(appearSymbolPair)
		}
	}
}

func(app *App) ProcessMexcAppearSymbolPair() {
	for {
		processMexcAppearSymbolPair(*app)
	}
}

func(app *App) ProcessMexcSymbolPairTicker(appearSymbolPair model.AppearSymbolPair) {
	supportRightSymbol := "USDT"
	if appearSymbolPair.Symbol1And2[1] != supportRightSymbol {
		log.WithFields(log.Fields{
			"appearSymbolPair": appearSymbolPair,
		}).Info("not support appearSymbolPair")
		return
	}

	for {
		processMexcSymbolPairTicker(*app, appearSymbolPair)
	}
}
