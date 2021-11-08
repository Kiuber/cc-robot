package cyaml

import (
	clog "cc-robot/core/tool/log"
	"fmt"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func LoadConfig(configFname string, out interface{}) {
	file, err := ioutil.ReadFile(fmt.Sprintf("config/%s", configFname))

	logger := clog.EventLog().With(zap.String(configFname, configFname))

	if err != nil {
		logger.Error("LoadConfig ReadFile err")
	}

	err = yaml.Unmarshal(file, out)
	if err != nil {
		logger.Error("LoadConfig Unmarshal err")
	}
}
