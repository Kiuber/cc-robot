package service

import (
	clog "cc-robot/core/tool/log"
	s_exchange "cc-robot/service/exchange"
	s_prime "cc-robot/service/prime"
)

type App struct {
	Prime    s_prime.Prime
	Exchange s_exchange.Exchange
}

func RunApp() *App {
	clog.EventLog.Info("run app")
	prime := s_prime.Main()
	exchange := s_exchange.Main()

	app := &App{
		Prime:    *prime,
		Exchange: *exchange,
	}
	return app
}
