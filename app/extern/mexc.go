package mexc

import (
	cboot "cc-robot/core/boot"
	chttp "cc-robot/core/tool/http"
	"cc-robot/model"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/mitchellh/mapstructure"
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

func Depth(symbol string, depth string) model.MexcAPIData {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("depth", depth)
	return mexcGetJson("market/depth", params)
}

func AccountInfo() model.MexcAPIData {
	mexcAPIData := mexcGetJson("account/info", nil)
	accountInfo := new(model.AccountInfo)
	mapstructure.Decode(mexcAPIData.RawPayload, &accountInfo)
	mexcAPIData.Payload = *accountInfo
	return mexcAPIData
}

func mexcGetJson(apiPath string, params url.Values) model.MexcAPIData {
	url := buildUrl(apiPath)
	header := buildHeader(params)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, _ := chttp.HttpGetJson(url, header)

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
	}

	return mexcAPIData
}

func buildUrl(apiPath string) string {
	return fmt.Sprintf("%s/%s", cboot.GV.Config.Api.Mexc.BaseURL, apiPath)
}

func buildHeader(params url.Values) http.Header {
	requestTime := strconv.FormatInt(time.Now().Unix() * 1000, 10)
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("ApiKey", cboot.GV.Config.Api.Mexc.AK)
	header.Add("Request-Time", requestTime)

	if params == nil {
		str := fmt.Sprintf("%s%s%s", cboot.GV.Config.Api.Mexc.AK, requestTime, params.Encode())
		header.Add("Signature", buildSignature(str))
	}
	return header
}

func buildSignature(data string) string {
	h := hmac.New(sha256.New, []byte(cboot.GV.Config.Api.Mexc.AS))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
