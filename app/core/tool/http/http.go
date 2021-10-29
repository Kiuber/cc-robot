package chttp

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

func HttpGetJson(url string, header http.Header) (data interface{}, err error) {
	resp, err := HttpGet(url, header)
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
	}).Info("HttpGetJson, unmarshal")

	return data, err
}

func HttpGet(url string, header http.Header) (resp *http.Response, err error) {
	header = mergeBasicHeader(header)

	client := http.Client{}
	req , err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("HttpGetJson, new request fail")
	}

	log.WithFields(log.Fields{
		"url": url,
		"header": header,
	}).Debug("HttpGet, request message")

	req.Header = header
	resp, err = client.Do(req)
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("HttpGetJson, new request do fail")
	}
	return resp, err
}

// HttpPost TODO: @qingbao, waiting for completion
func HttpPost(url, contentType string, body io.Reader) (resp *http.Response, err error) {
	return http.Post(url, contentType, body)
}

func mergeBasicHeader(header http.Header) http.Header {
	basicHeader := BasicHeader()
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

func BasicHeader() http.Header {
	header := http.Header{}
	header.Add("Identity", "hi")
	return header
}