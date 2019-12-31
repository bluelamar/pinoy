package config

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

type PinoyConfig struct {
	DbType                    string // supported type, currently either "couchdb" or "mongodb"
	DbUrl                     string
	DbName                    string
	DbPort                    int
	DbUser                    string
	DbPwd                     string // encrypted - user defined encryption
	DbCommTimeout             int    // communication timeout in seconds
	LogOutput                 string // stdout, file
	LogFile                   string // path name if LogOutput == "file"
	CookieSecret              string // if set, use this as the cookie secret rather than from the DB
	CookieSecretDbEntity      string // if set, retrieve cookie secret from thid db entity
	StatsMonitorInterval      int    // minutes, report server stats
	RoomStatusMonitorInterval int    // minutes, check room status for rooms to be checked out
	CheckoutGracePeriod       int    // minutes, grace period for customer to checkout
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
	cfg.NormalizeConfig()
	return &cfg, nil
}

// NormalizeConfig ensures all fields have sane values
func (cfg *PinoyConfig) NormalizeConfig() {
	if cfg.DbType == "" {
		cfg.DbType = "mongodb"
	}
	if cfg.DbUrl == "" {
		cfg.DbUrl = "127.0.0.1"
	}
	if cfg.DbName == "" {
		cfg.DbName = "pinoy"
	}
	if cfg.DbPort == 0 {
		cfg.DbPort = 27017
	}

	if cfg.DbCommTimeout == 0 { // communication timeout in seconds
		cfg.DbCommTimeout = 5
	}
	if cfg.LogOutput == "" { // stdout, file
		cfg.LogOutput = "stdout"
	}

	if cfg.StatsMonitorInterval == 0 { // minutes, report server stats
		cfg.StatsMonitorInterval = 60
	}

	if cfg.RoomStatusMonitorInterval == 0 { // minutes, check room status for rooms to be checked out
		cfg.RoomStatusMonitorInterval = 2
	}
	if cfg.CheckoutGracePeriod == 0 { // minutes, grace period for customer to checkout
		cfg.CheckoutGracePeriod = 15
	}
}
