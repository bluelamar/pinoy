package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type PinoyConfig struct {
	DbUrl                     string
	DbName                    string
	DbPort                    int
	DbUser                    string
	DbPwd                     string
	DbCommTimeout             int    // communication timeout in seconds
	LogOutput                 string // stdout, file
	LogFile                   string // path name if LogOutput == "file"
	CookieSecret              string // if set, use this as the cookie secret rather than from the DB
	CookieSecretDbEntity      string // if set, retrieve cookie secret from thid db entity
	StatsMonitorInterval      int    // minutes, report server stats
	RoomStatusMonitorInterval int    // minutes, check room status for rooms to be checked out
}

func LoadConfig(fpath string) (*PinoyConfig, error) {
	content, err := ioutil.ReadFile(fpath)
	if err != nil {
		log.Println("Config:ERROR: Failed to read config file:", err)
		return nil, err
	}

	var cfg PinoyConfig
	err = json.Unmarshal(content, &cfg)
	if err != nil {
		log.Println("Config:ERROR: Failed to unmarshal config file:", err)
	}
	return &cfg, nil
}
