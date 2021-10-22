package mexc

import (
	chttp "cc-robot/core/tool/http"
	"cc-robot/model"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const baseUrl = "https://www.mexc.com/open/api/v2"

func Ping(ctx model.Context) interface{} {
	return mexcGetJson(ctx, "common/ping", nil)
}

func Timestamp(ctx model.Context) interface{} {
	return mexcGetJson(ctx, "common/timestamp", nil)
}

func Symbols(ctx model.Context) interface{} {
	return mexcGetJson(ctx, "market/api_symbols", nil)
}

func Depth(ctx model.Context, symbol string, depth string) interface{} {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("depth", depth)
	return mexcGetJson(ctx, "market/depth", params)
}

func AccountInfo(ctx model.Context) interface{} {
	return mexcGetJson(ctx, "account/info", nil)
}

func mexcGetJson(ctx model.Context, apiPath string, params url.Values) interface{} {
	url := buildUrl(apiPath)
	header := buildHeader(ctx, params)

	if params != nil {
		url = fmt.Sprintf("%s?%s", url, params.Encode())
	}
	resp, _ := chttp.HttpGetJson(url, header)
	return resp
}

func buildUrl(apiPath string) string {
	return fmt.Sprintf("%s/%s", baseUrl, apiPath)
}

func buildHeader(ctx model.Context, params url.Values) http.Header {
	requestTime := strconv.FormatInt(time.Now().Unix() * 1000, 10)
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("ApiKey", ctx.Config.Api.Mexc.AK)
	header.Add("Request-Time", requestTime)

	if params == nil {
		str := fmt.Sprintf("%s%s%s", ctx.Config.Api.Mexc.AK, requestTime, params.Encode())
		header.Add("Signature", buildSignature(ctx, str))
	}
	return header
}

func buildSignature(ctx model.Context, data string) string {
	h := hmac.New(sha256.New, []byte(ctx.Config.Api.Mexc.AS))
	h.Write([]byte(data))
	sha := hex.EncodeToString(h.Sum(nil))
	return sha
}
