package cjson

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/pretty"
	"io/ioutil"
	"os"
)

func Pretty(v interface{}) string {
	json, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(pretty.Pretty(json))
}

func UnmarshalFromFile(name string, v interface{}) {
	file, err := os.Open(name)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

	byteValue, _ := ioutil.ReadAll(file)
	json.Unmarshal(byteValue, &v)
}
