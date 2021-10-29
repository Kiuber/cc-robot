package cjson

import (
	"encoding/json"
	"github.com/tidwall/pretty"
)

func Pretty(v interface{}) string {
	json, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(pretty.Pretty(json))
}
