package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/prometheus/client_golang/prometheus"
)

type HomekitBridge struct {
	ListenAddress       string       `json:"ListenAddress"`
	UserName            string       `json:"UserName"`
	Password            string       `json:"Password"`
	AccessoryList       []Accessorys `json:"AccessoryList"`
	MetricListenAddress string       `json:"MetricListenAddress"`
	metricstatus        *prometheus.GaugeVec
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
	if len(s.MetricListenAddress) == 0 {
		s.MetricListenAddress = "0.0.0.0:7080"
	}
	s.metricstatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "smarthome",
			Subsystem: "homekit",
			Name:      "bridge",
			Help:      "homekit bridge status.",
		},
		[]string{"serialnumber", "sensor"},
	)
	// Register status
	prometheus.Register(s.metricstatus)
	return s, nil
}

func (hb *HomekitBridge) Tasks() {
	for _, v := range hb.AccessoryList {
		ac := v
		ac.hb = hb
		ac.dataChannel = make(chan float64)
		go ac.Task()
	}
}
