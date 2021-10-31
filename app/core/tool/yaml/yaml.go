package cyaml

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func LoadConfig(configFname string, out interface{}) {
	file, err := ioutil.ReadFile(fmt.Sprintf("config/%s", configFname))

	logger := log.WithFields(log.Fields{configFname: configFname})

	if err != nil {
		logger.Error("LoadConfig ReadFile err")
	}

	err = yaml.Unmarshal(file, out)
	if err != nil {
		logger.Error("LoadConfig Unmarshal err")
	}
}
