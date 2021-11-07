package mexc

import (
	"bytes"
	cboot "cc-robot/core/boot"
	chttp "cc-robot/core/tool/http"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func Ping() model.MexcAPIData {
	return mexcGetJson("common/ping", nil)
}

func Timestamp() model.MexcAPIData {
	return mexcGetJson("common/timestamp", nil)
}

func SupportSymbolPair() model.MexcAPIData {
	mexcAPIData := mexcGetJson("market/api_symbols", nil)

	supportSymbols := new(model.SupportSymbolPair)
	mapstructure.Decode(mexcAPIData.RawPayload, &supportSymbols)
	supportSymbols.Exchange = "mexc"
	mexcAPIData.Payload = *supportSymbols
	return mexcAPIData
}

func SymbolPairInfoList() model.MexcAPIData {
	mexcAPIData := mexcGetJson("market/symbols", nil)

	SymbolList := new(model.SymbolPairInfoList)
	mapstructure.Decode(mexcAPIData.RawPayload, &SymbolList)
	mexcAPIData.Payload = *SymbolList
	return mexcAPIData
}

func SymbolPairTickerInfo(symbolPair string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	mexcAPIData := mexcGetJson("market/ticker", params)

	symbolPairTickerInfo := &model.SymbolPairTickerInfo{}
	mapstructure.Decode(mexcAPIData.RawPayload, symbolPairTickerInfo)
	mexcAPIData.Payload = *symbolPairTickerInfo
	return mexcAPIData
}

func DepthInfo(symbolPair string, depth string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("depth", depth)
	mexcAPIData := mexcGetJson("market/depth", params)

	depthInfo := &model.DepthInfo{}
	mapstructure.Decode(mexcAPIData.RawPayload, depthInfo)
	mexcAPIData.Payload = *depthInfo
	return mexcAPIData
}

func CreateOrder(order model.Order) model.MexcAPIData {
	if !cboot.GV.Config.Api.Mexc.AllowTrade {
		return model.MexcAPIData{}
	}
	json, _ := json.Marshal(order)
	mexcAPIData := mexcPostJson("order/place", json)
	return mexcAPIData
}

func OrderList(symbolPair string, tradeType string, states string, limit string, startTime string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("trade_type", tradeType)
	params.Set("states", states)
	params.Set("limit", limit)
	params.Set("start_time", startTime)
	mexcAPIData := mexcGetJson("order/list", params)

	orderList := new(model.OrderList)
	mapstructure.Decode(mexcAPIData.RawPayload, &orderList)
	mexcAPIData.Payload = *orderList
	return mexcAPIData
}

func CancelOrder(symbolPair string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	mexcAPIData := mexcDeleteJson("order/cancel_by_symbol", params)
	return mexcAPIData
}

func AccountInfo() model.MexcAPIData {
	mexcAPIData := mexcGetJson("account/info", nil)
	accountInfo := new(model.AccountInfo)
	mapstructure.Decode(mexcAPIData.RawPayload, &accountInfo)
	mexcAPIData.Payload = *accountInfo
	return mexcAPIData
}

func KLine(symbolPair string, interval string, startTime string, limit string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("interval", interval)
	params.Set("start_time", startTime)
	params.Set("limit", limit)
	mexcAPIData := mexcGetJson("market/kline", params)
	kLine := &[][]int{}
	mapstructure.Decode(mexcAPIData.RawPayload, &kLine)
	mexcAPIData.Payload = *kLine
	return mexcAPIData
}

func mexcGetJson(apiPath string, params url.Values) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(params, nil)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, _ := chttp.HttpGetJson(url, header)
	return processResp(url, resp)
}

func mexcPostJson(apiPath string, data []byte) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(nil, data)

	resp, _ := chttp.HttpPostJson(url, header, bytes.NewBuffer(data))
	return processResp(url, resp)
}

func mexcDeleteJson(apiPath string, params url.Values) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(params, nil)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, _ := chttp.HttpDeleteJson(url, header, nil)
	return processResp(url, resp)
}

func processResp(url string, resp interface{}) model.MexcAPIData {
	var mexcResp model.MexcResp
	cfg := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   &mexcResp,
		TagName:  "json",
	}
	decoder, _ := mapstructure.NewDecoder(cfg)
	decoder.Decode(resp)

	mexcAPIData := model.MexcAPIData{Payload: ""}
	if mexcResp.Code == 200 {
		mexcAPIData.OK = true
		mexcAPIData.RawPayload = mexcResp.Data
	} else {
		clog.EventLog().WithFields(logrus.Fields{
			"url":  url,
			"resp": resp,
		}).Error("request mexc API failed")
	}

	return mexcAPIData
}

func buildUrl(apiPath string) string {
	return fmt.Sprintf("%s/%s", cboot.GV.Config.Api.Mexc.BaseURL, apiPath)
}

func buildHeader(params url.Values, data []byte) http.Header {
	requestTime := strconv.FormatInt(time.Now().Unix()*1000, 10)
	header := http.Header{}
	header.Set("Content-Type", "application/json")
	header.Set("ApiKey", cboot.GV.Config.Api.Mexc.AK)
	header.Set("Request-Time", requestTime)

	signPrefixStr := fmt.Sprintf("%s%s", cboot.GV.Config.Api.Mexc.AK, requestTime)
	if data == nil {
		signStr := fmt.Sprintf("%s%s", signPrefixStr, params.Encode())
		header.Set("Signature", buildSignature(signStr))
	} else {
		signStr := fmt.Sprintf("%s%s", signPrefixStr, string(data))
		header.Set("Signature", buildSignature(signStr))
	}
	return header
}

func buildSignature(data string) string {
	h := hmac.New(sha256.New, []byte(cboot.GV.Config.Api.Mexc.AS))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
