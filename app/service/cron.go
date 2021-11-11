package service

import (
	ccron "cc-robot/core/tool/cron"
	clog "cc-robot/core/tool/log"
)

func RunCron(app *App) {
	clog.EventLog.Info("run cron")
	c := ccron.New()
	ccron.AddFunc(c, "10 * * * *", func() {
		app.FetchAndUpsertAPISupportSymbolPairs()
	}, true)
	ccron.AddFunc(c, "10 * * * *", func() {
		app.CheckAndAlarmForSymbolPairs()
	}, true)
	c.Start()
}
