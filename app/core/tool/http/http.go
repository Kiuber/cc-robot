package chttp

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

func HttpGetJson(url string, header http.Header) (data interface{}, err error) {
	resp, err := httpGet(url, header)
	return jsonifyResp(resp)
}

func HttpPostJson(url string, header http.Header, body io.Reader) (data interface{}, err error) {
	resp, err := httpPost(url, header, body)
	return jsonifyResp(resp)
}

func HttpDeleteJson(url string, header http.Header, body io.Reader) (data interface{}, err error) {
	resp, err := httpDelete(url, header, body)
	return jsonifyResp(resp)
}

func httpGet(url string, header http.Header) (resp *http.Response, err error) {
	req, err := buildRequest("GET", url, header, nil)
	resp, err = doRequest(req)
	return resp, err
}

func httpPost(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := buildRequest("POST", url, header, body)
	resp, err = doRequest(req)
	return resp, err
}

func httpDelete(url string, header http.Header, body io.Reader) (resp *http.Response, err error) {
	req, err := buildRequest("DELETE", url, header, body)
	resp, err = doRequest(req)
	return resp, err
}

func buildRequest(method string, url string, header http.Header, body io.Reader) (resp *http.Request, err error) {
	req , err := http.NewRequest(method, url, body)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("new request failed")
	}
	req.Header = mergeBasicHeader(req, header)
	return req, nil
}

func doRequest(req *http.Request) (resp *http.Response, err error) {
	logger := log.WithFields(log.Fields{
		"method": req.Method,
		"url": req.URL,
		"header": req.Header,
		"data": req.Body,
	})
	logger.Info("request info")

	client := http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		logger.WithFields(log.Fields{
			"err": err,
		}).Error("request failed")
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

func jsonifyResp(resp *http.Response) (data interface{}, err error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.WithFields(log.Fields{
			"resp": resp,
			"err": err,
		}).Error("HttpGetJson, read response")
	}

	bodyString := string(bodyBytes)
	err = json.Unmarshal(bodyBytes, &data)

	log.WithFields(log.Fields{
		"bodyString": bodyString,
	}).Debug("HttpPostJson, unmarshal")
	return data, err
}

func basicHeader() http.Header {
	hostname, _ := os.Hostname()
	header := http.Header{}
	header.Set("Extra-Request-Hostname", hostname)
	return header
}