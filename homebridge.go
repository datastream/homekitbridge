package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/brutella/hap"
	"github.com/brutella/hap/accessory"
)

type HomekitBridge struct {
	ListenAddress string       `json:"ListenAddress"`
	UserName      string       `json:"UserName"`
	Password      string       `json:"Password"`
	AccessoryList []Accessorys `json:"AccessoryList"`

	Name         string `json:"name"`
	SerialNumber string `json:"serialNumber"`
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Pin          string `json:"Pin"`

	MetricListenAddress string `json:"MetricListenAddress"`
}

func ReadConfig(file string) (*HomekitBridge, error) {
	configFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	config, err := io.ReadAll(configFile)
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

func (hb *HomekitBridge) Tasks(ctx context.Context) {
	var accessorys []*accessory.A
	for _, v := range hb.AccessoryList {
		ac := v
		ac.hb = hb
		ac.dataChannel = make(chan float64)
		ac.exitChan = make(chan int)
		acc := ac.Task()
		accessorys = append(accessorys, acc)
	}
	if len(accessorys) < 1 {
		return
	}
	fs := hap.NewFsStore("./bridges")
	var server *hap.Server
	var err error
	info := accessory.Info{
		Name:         hb.Name,
		SerialNumber: hb.SerialNumber,
		Manufacturer: hb.Manufacturer,
		Model:        hb.Model,
	}
	basicAccess := accessory.NewBridge(info)
	if len(accessorys) > 1 {
		server, err = hap.NewServer(fs, basicAccess.A, accessorys...)
	} else {
		server, err = hap.NewServer(fs, basicAccess.A)
	}
	if err != nil {
		log.Panic(err)
	}
	server.Pin = hb.Pin
	go server.ListenAndServe(ctx)
}
