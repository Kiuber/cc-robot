package cid

import (
	clog "cc-robot/core/tool/log"
	"github.com/sirupsen/logrus"
	"github.com/sony/sonyflake"
	"strconv"
)

func UniuqeId() string {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		clog.EventLog().WithFields(logrus.Fields{"err": err}).Error("nextID")
	}
	return strconv.FormatUint(id, 10)
}
