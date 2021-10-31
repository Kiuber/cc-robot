package service

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func RunApp() {
	log.Info("run app")
	initLogic()
}

func initLogic() {
	for {
		HandleMexcSymbolPair()
		time.Sleep(time.Second * 10)
	}
}
