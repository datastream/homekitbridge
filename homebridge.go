package main

import (
	"encoding/json"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"os"
)

type HomekitBridge struct {
	ListenAddress string `json:"ListenAddress"`
	UserName      string `json:"UserName"`
	Password      string `json:"Password"`
	cache         *cache.Cache
	Topic         string
	AccessoryList []Accessorys `json:"AccessoryList"`
}

func ReadConfig(file string) (*HomekitBridge, error) {
	configFile, err := os.Open(file)
	config, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	var s *HomekitBridge
	if err := json.Unmarshal(config, &s); err != nil {
		return nil, err
	}
	return s, nil
}

func (hb *HomekitBridge) Tasks() {
	for _, v := range hb.AccessoryList {
		ac := v
		ac.hb = hb
		go ac.Task()
	}
}
