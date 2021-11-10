package chttp

import (
	cid "cc-robot/core/tool/id"
	clog "cc-robot/core/tool/log"
	"encoding/json"
	"go.uber.org/zap"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

const ExtraRequestIdField = "Extra-Request-Id"

func HttpGetJson(url string, header http.Header) (data interface{}, err error) {
	resp, err, req := httpGet(url, header)
	return jsonifyResp(resp, req)
}

func HttpPostJson(url string, header http.Header, body io.Reader) (data interface{}, err error) {
	resp, err, req := httpPost(url, header, body)
	return jsonifyResp(resp, req)
}

func HttpDeleteJson(url string, header http.Header, body io.Reader) (data interface{}, err error) {
	resp, err, req := httpDelete(url, header, body)
	return jsonifyResp(resp, req)
}

func httpGet(url string, header http.Header) (resp *http.Response, err error, req *http.Request) {
	req, err = buildRequest("GET", url, header, nil)
	resp, err = doRequest(req)
	return resp, err, req
}

func httpPost(url string, header http.Header, body io.Reader) (resp *http.Response, err error, req *http.Request) {
	req, err = buildRequest("POST", url, header, body)
	resp, err = doRequest(req)
	return resp, err, req
}

func httpDelete(url string, header http.Header, body io.Reader) (resp *http.Response, err error, req *http.Request) {
	req, err = buildRequest("DELETE", url, header, body)
	resp, err = doRequest(req)
	return resp, err, req
}

func buildRequest(method string, url string, header http.Header, body io.Reader) (resp *http.Request, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		clog.VerboseLog.With(zap.String("err", err.Error())).Error("new request failed")
	}
	req.Header = mergeBasicHeader(req, header)
	return req, nil
}

func doRequest(req *http.Request) (resp *http.Response, err error) {
	logger := clog.VerboseLog.With(
		zap.String("method", req.Method),
		zap.String("url", req.URL.String()),
		zap.Reflect("header", req.Header),
		zap.Reflect("data", req.Body),
	)
	logger.Info("request info")

	client := http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		logger.With(zap.String("err", err.Error())).Error("request failed")
	}

	return resp, err
}

func mergeBasicHeader(req *http.Request, header http.Header) http.Header {
	basicHeader := basicHeader()
	if header == nil {
		return basicHeader
	}

	for key, values := range basicHeader {
		for _, value := range values {
			header.Add(key, value)
		}
	}
	return header
}

func jsonifyResp(resp *http.Response, req *http.Request) (data interface{}, err error) {
	if resp == nil {
		clog.VerboseLog.Error("jsonify, response nil")
		return new(interface{}), err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		clog.VerboseLog.With(
			zap.String("err", err.Error()),
		).Error("jsonify, read response")
		return new(interface{}), err
	}

	respStr := string(bodyBytes)
	err = json.Unmarshal(bodyBytes, &data)

	logger := clog.VerboseLog.With(
		zap.String("respStr", respStr),
		zap.String(ExtraRequestIdField, req.Header.Get(ExtraRequestIdField)),
		zap.Reflect("err", err),
	)
	if err != nil {
		logger.Error("jsonify failed")
	} else {
		logger.Info("jsonify succeed")
	}
	return data, err
}

func basicHeader() http.Header {
	hostname, _ := os.Hostname()
	header := http.Header{}
	header.Set("Extra-Request-Hostname", hostname)
	header.Set(ExtraRequestIdField, cid.UniuqeId())
	return header
}
