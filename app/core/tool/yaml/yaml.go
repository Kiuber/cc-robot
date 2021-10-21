package cyaml

import (
	"cc-robot/module"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

// LoadConfig TODO: @qingbao, adjust paramter type and return type
func LoadConfig() module.ApiConfig {
	out := new(module.ApiConfig)
	file, err := ioutil.ReadFile("config/config.yaml")

	if err != nil {
		log.WithFields(log.Fields{
			"file": file,
			"err": err,
		}).Error("LoadConfig ReadFile err")
	}
	err = yaml.Unmarshal(file, out)
	if err != nil {
		log.WithFields(log.Fields{
			"file": file,
			"err": err,
		}).Error("LoadConfig Unmarshal err")
	}
	return *out
}
