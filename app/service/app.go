package service

import (
	cinfra "cc-robot/core/tool/infra"
	"cc-robot/model"
	"fmt"
	log "github.com/sirupsen/logrus"
)

type App struct {
	symbolPairCh chan model.AppearSymbolPair
}

func RunApp() {
	log.Info("run app")
	app := initApp()
	app.initLogic()
}

func initApp() *App {
	app := &App{
		symbolPairCh: make(chan model.AppearSymbolPair),
	}
	return app
}

func(app *App) initLogic() {
	go app.HandleMexcSymbolPair()
	for {
		select {
		case appearSymbolPair := <-app.symbolPairCh:
			cinfra.GiantEventText(fmt.Sprintf("%s symbol pair appear %s", appearSymbolPair.Exchange, appearSymbolPair.SymbolPair))
		}
	}
}
