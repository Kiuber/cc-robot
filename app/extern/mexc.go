package emexc

import (
	"bytes"
	cboot "cc-robot/core/boot"
	chttp "cc-robot/core/tool/http"
	clog "cc-robot/core/tool/log"
	"cc-robot/model"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Ping() model.MexcAPIData {
	return mexcGetJson(nil, "common/ping", nil)
}

func Timestamp() model.MexcAPIData {
	return mexcGetJson(nil, "common/timestamp", nil)
}

func SupportSymbolPair(ctx context.Context) model.MexcAPIData {
	mexcAPIData := mexcGetJson(ctx, "market/api_symbols", nil)

	supportSymbols := new(model.SupportSymbolPair)
	mapstructure.Decode(mexcAPIData.RawPayload, &supportSymbols)
	mexcAPIData.Payload = *supportSymbols
	return mexcAPIData
}

func SymbolPairInfoList() model.MexcAPIData {
	mexcAPIData := mexcGetJson(nil, "market/symbols", nil)

	SymbolList := new(model.SymbolPairInfoList)
	mapstructure.Decode(mexcAPIData.RawPayload, &SymbolList)
	mexcAPIData.Payload = *SymbolList
	return mexcAPIData
}

func SymbolPairTickerInfo(symbolPair string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	mexcAPIData := mexcGetJson(nil, "market/ticker", params)

	symbolPairTickerInfo := &model.SymbolPairTickerInfo{}
	mapstructure.Decode(mexcAPIData.RawPayload, symbolPairTickerInfo)
	mexcAPIData.Payload = *symbolPairTickerInfo
	return mexcAPIData
}

func DepthInfo(ctx context.Context, symbolPair string, depth string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("depth", depth)
	mexcAPIData := mexcGetJson(ctx, "market/depth", params)

	depthInfo := &model.DepthInfo{}
	mapstructure.Decode(mexcAPIData.RawPayload, depthInfo)
	mexcAPIData.Payload = *depthInfo
	return mexcAPIData
}

func CreateOrder(ctx context.Context, order model.Order) model.MexcAPIData {
	if !cboot.GV.Config.Api.Mexc.AllowTrade {
		return model.MexcAPIData{}
	}
	json, _ := json.Marshal(order)
	mexcAPIData := mexcPostJson(ctx, "order/place", json)
	return mexcAPIData
}

func OrderList(ctx context.Context, symbolPair string, tradeType string, states string, limit string, startTime string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("trade_type", tradeType)
	params.Set("states", states)
	params.Set("limit", limit)
	params.Set("start_time", startTime)
	mexcAPIData := mexcGetJson(ctx, "order/list", params)

	orderList := new(model.OrderList)
	mapstructure.Decode(mexcAPIData.RawPayload, &orderList)
	mexcAPIData.Payload = *orderList
	return mexcAPIData
}

func CancelOrder(ctx context.Context, symbolPair string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	mexcAPIData := mexcDeleteJson(ctx, "order/cancel_by_symbol", params)
	return mexcAPIData
}

func AccountInfo(ctx context.Context) model.MexcAPIData {
	mexcAPIData := mexcGetJson(ctx, "account/info", nil)
	accountInfo := new(model.AccountInfo)
	mapstructure.Decode(mexcAPIData.RawPayload, &accountInfo)
	mexcAPIData.Payload = *accountInfo
	return mexcAPIData
}

func KLine(ctx context.Context, symbolPair string, interval string, startTime string, limit string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	params.Set("interval", interval)
	params.Set("start_time", startTime)
	params.Set("limit", limit)
	mexcAPIData := mexcGetJson(ctx, "market/kline", params)
	kLine := &[]interface{}{}
	mapstructure.Decode(mexcAPIData.RawPayload, &kLine)
	mexcAPIData.Payload = *kLine
	return mexcAPIData
}

func SymbolPairInfo(ctx context.Context, symbolPair string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbolPair)
	mexcAPIData := mexcGetJson(ctx, "api/platform/spot/market/symbol", params)
	symbolPairInfo := new(model.SymbolPairInfo)
	mapstructure.Decode(mexcAPIData.RawPayload, &symbolPairInfo)
	symbolPairInfo.SymbolPair = symbolPair
	mexcAPIData.Payload = *symbolPairInfo
	return mexcAPIData
}

func SymbolIntroduceInfo(ctx context.Context, symbol string) model.MexcAPIData {
	params := url.Values{}
	params.Set("currency", symbol)
	mexcAPIData := mexcGetJson(ctx, "api/platform/spot/market/introduce", params)
	symbolIntroduceInfo := new(model.SymbolIntroduceInfo)
	mapstructure.Decode(mexcAPIData.RawPayload, &symbolIntroduceInfo)
	symbolIntroduceInfo.Symbol = symbol
	mexcAPIData.Payload = *symbolIntroduceInfo
	return mexcAPIData
}

func mexcGetJson(ctx context.Context, apiPath string, params url.Values) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(params, nil)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, err := chttp.HttpGetJson(ctx, url, header)
	return processResp(ctx, url, resp, err)
}

func mexcPostJson(ctx context.Context, apiPath string, data []byte) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(nil, data)

	resp, err := chttp.HttpPostJson(ctx, url, header, bytes.NewBuffer(data))
	return processResp(ctx, url, resp, err)
}

func mexcDeleteJson(ctx context.Context, apiPath string, params url.Values) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(params, nil)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, err := chttp.HttpDeleteJson(ctx, url, header, nil)
	return processResp(ctx, url, resp, err)
}

func processResp(ctx context.Context, url string, resp interface{}, err error) model.MexcAPIData {
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
	} else if err != nil {
		mexcAPIData.Msg = fmt.Sprintf("parse resp failed, error: %s", err.Error())
		clog.WithCtxEventLog(ctx).With(
			zap.String("url", url),
			zap.Reflect("resp", resp),
		).Error("request mexc API failed")
	} else if mexcResp.Code != 200 {
		mexcAPIData.Msg = mexcResp.Msg
	}

	return mexcAPIData
}

func buildUrl(apiPath string) string {
	if strings.HasPrefix(apiPath, "api/") {
		return fmt.Sprintf("%s/%s", cboot.GV.Config.Api.Mexc.BaseURL, apiPath)
	} else {
		return fmt.Sprintf("%s/%s/%s", cboot.GV.Config.Api.Mexc.BaseURL, "open/api/v2", apiPath)
	}
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
