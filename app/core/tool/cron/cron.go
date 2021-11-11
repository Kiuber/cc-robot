package ccron

import (
	"github.com/robfig/cron/v3"
)

func New() *cron.Cron {
	c := cron.New()
	return c
}

func AddFunc(c *cron.Cron, spec string, cmd func(), execImmediately bool) {
	c.AddFunc(spec, cmd)
	if execImmediately {
		cmd()
	}
}
