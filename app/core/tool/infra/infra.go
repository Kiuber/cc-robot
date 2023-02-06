package cinfra

import (
	cboot "cc-robot/core/boot"
	chttp "cc-robot/core/tool/http"
	"cc-robot/model"
	"context"
	"fmt"
	"net/url"
)

func GiantEventText(ctx context.Context, msg string) {
	var notifyGroup string
	if cboot.GV.IsDev {
		notifyGroup = model.EventNotifyGroupDev
	} else {
		notifyGroup = model.EventNotifyGroupProd
	}
	params := url.Values{}
	params.Add("notify_group", notifyGroup)
	params.Add("notify_key", "cc_robot")
	params.Add("msg", msg)
	url := fmt.Sprintf("%s?%s", cboot.GV.Config.Infra.EventNU.URL, params.Encode())
	chttp.HttpGetJson(ctx, url, nil)
}
