package service

import (
	log "github.com/sirupsen/logrus"
	"time"
)

func RunApp() {
	initLogic()
}

func initLogic() {
	for {
		log.Info("RunApp")
		HandleMexcSymbolPair()
		time.Sleep(time.Second * 3)
	}
}
