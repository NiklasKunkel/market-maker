package config

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"github.com/niklaskunkel/market-maker/logger"
	"github.com/sirupsen/logrus"
)

//Globals
var log = logger.InitLogger()

type Auth struct {
	Key		string	`json:"apiKey"`
	Secret	string 	`json:"apiSecret"`
}

type Config struct {
	LogPath		string 	`json:"logPath"`
	SetzerPath	string 	`json:"setzerPath"`
}

func LoadCredentials(credentials *Auth) {
	LoadFile(credentials, "credentials.json")
	return
}

func LoadConfig(config *Config) {
	LoadFile(config, "config.json")
	return
}

func LoadFile(filetype interface{}, filename string) {
	goPath, ok := os.LookupEnv("GOPATH")
	if ok != true {
		log.WithFields(logrus.Fields{"function": "LoadFile"}).Fatal("$GOPATH Env Variable not set")
	}
	filePath := goPath + "/src/github.com/niklaskunkel/market-maker/" + filename
	raw, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.WithFields(logrus.Fields{"function": "LoadFile", "path": filePath, "error": err.Error()}).Fatal("Unable to read file")
	}
	err = json.Unmarshal(raw, filetype)
	if err != nil {
		log.WithFields(logrus.Fields{"function": "LoadFile", "file": filename, "json": raw, "error": err.Error()}).Fatal("Unable to parse JSON")
	}
	return
}