package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/service"
	"log"
	"strconv"
	"time"
)

type Accessorys struct {
	Key           string `json:"key"`
	Name          string `json:"name"`
	SerialNumber  string `json:"serialNumber"`
	Manufacturer  string `json:"manufacturer"`
	Model         string `json:"model"`
	Pin           string `json:"pin"`
	AccessoryType string `json:"accessoryType"`
}

type HumiditySensor struct {
	*accessory.Accessory

	HumiditySensor *service.HumiditySensor
}

func NewHumiditySensor(info accessory.Info, cur, min, max, steps float64) *HumiditySensor {
	acc := HumiditySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeHumidifer)
	acc.HumiditySensor = service.NewHumiditySensor()
	acc.HumiditySensor.CurrentRelativeHumidity.SetValue(cur)
	acc.HumiditySensor.CurrentRelativeHumidity.SetMinValue(min)
	acc.HumiditySensor.CurrentRelativeHumidity.SetMaxValue(max)
	acc.HumiditySensor.CurrentRelativeHumidity.SetStepValue(steps)

	acc.AddService(acc.HumiditySensor.Service)

	return &acc
}
func (ac *Accessorys) Task() {
	info := accessory.Info{
		Name:         ac.Name,
		SerialNumber: ac.SerialNumber,
		Manufacturer: ac.Manufacturer,
		Model:        ac.Model,
	}
	switch ac.AccessoryType {
	case "TemperatureSensor":
		acc := accessory.NewTemperatureSensor(info, 5, -100, 50, 0.1)
		config := hc.Config{Pin: ac.Pin}
		t, err := hc.NewIPTransport(config, acc.Accessory)
		if err != nil {
			log.Panic(err)
		}

		hc.OnTermination(func() {
			t.Stop()
		})
		go t.Start()
		for {
			value, found := homekitBridge.cache.Get(ac.Key)
			if !found {
				log.Println("bad key", ac.Key)
				time.Sleep(time.Second * 60)
				continue
			}
			temp, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Println("bad value", value)
				time.Sleep(time.Second * 60)
				continue
			}
			log.Println("get value", value, info)
			acc.TempSensor.CurrentTemperature.SetValue(temp)
			time.Sleep(time.Second * 60)
		}
	case "HumiditySensor":
		acc := NewHumiditySensor(info, 5, 0, 200, 0.1)
		config := hc.Config{Pin: ac.Pin}
		t, err := hc.NewIPTransport(config, acc.Accessory)
		if err != nil {
			log.Panic(err)
		}

		hc.OnTermination(func() {
			t.Stop()
		})
		go t.Start()
		for {
			value, found := homekitBridge.cache.Get(ac.Key)
			if !found {
				log.Println("bad key", ac.Key)
				time.Sleep(time.Second * 60)
				continue
			}
			hum, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Println("bad value", value)
				time.Sleep(time.Second * 60)
				continue
			}
			log.Println("get value", value, info)
			acc.HumiditySensor.CurrentRelativeHumidity.SetValue(hum)
			time.Sleep(time.Second * 60)
		}
	case "Switch":
		log.Println("test")
	}
}
