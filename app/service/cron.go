package service

import (
	ccron "cc-robot/core/tool/cron"
	clog "cc-robot/core/tool/log"
)

func RunCron(app *App) {
	clog.EventLog.Info("run cron")
	c := ccron.New()
	ccron.AddFunc(c, "*/20 * * * *", func() {
		app.Exchange.SaveAPISupportSymbolPairsOfAllExchanges()
	}, true)
	ccron.AddFunc(c, "*/20 * * * *", func() {
		app.Prime.CheckAndAlarmSymbolPairsOfAllExchanges()
	}, true)
	ccron.AddFunc(c, "*/20 * * * *", func() {
		app.Prime.TryUpdatePrimeSymbolPair()
	}, true)
	c.Start()
}
