package cid

import (
	clog "cc-robot/core/tool/log"
	"github.com/sony/sonyflake"
	"go.uber.org/zap"
	"strconv"
)

var flake *sonyflake.Sonyflake

func UniuqeId() string {
	if flake == nil {
		flake = sonyflake.NewSonyflake(sonyflake.Settings{})
	}
	id, err := flake.NextID()
	if err != nil {
		clog.EventLog.With(zap.String("err", err.Error())).Error("nextID")
	}
	return strconv.FormatUint(id, 10)
}
