package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type HomekitBridge struct {
	ListenAddress       string       `json:"ListenAddress"`
	UserName            string       `json:"UserName"`
	Password            string       `json:"Password"`
	AccessoryList       []Accessorys `json:"AccessoryList"`
	MetricListenAddress string       `json:"MetricListenAddress"`
}

func ReadConfig(file string) (*HomekitBridge, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	config, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	var s *HomekitBridge
	if err := json.Unmarshal(config, &s); err != nil {
		return nil, err
	}
	if len(s.MetricListenAddress) == 0 {
		s.MetricListenAddress = "0.0.0.0:7080"
	}
	return s, nil
}

func (hb *HomekitBridge) Tasks() {
	for _, v := range hb.AccessoryList {
		ac := v
		ac.hb = hb
		ac.dataChannel = make(chan float64)
		ac.exitChan = make(chan int)
		go ac.Task()
	}
}
