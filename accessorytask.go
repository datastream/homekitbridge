package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
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

type AirQualitySensorService struct {
	*service.Service
	AirQuality   *characteristic.AirQuality
	PM2_5Density *characteristic.PM2_5Density
}

type AirQualitySensor struct {
	*accessory.Accessory

	AirQualitySensor *AirQualitySensorService
}

func NewAirQualitySensorService() *AirQualitySensorService {
	svc := AirQualitySensorService{}
	svc.Service = service.New(service.TypeAirQualitySensor)
	svc.AirQuality = characteristic.NewAirQuality()
	svc.AddCharacteristic(svc.AirQuality.Characteristic)
	svc.PM2_5Density = characteristic.NewPM2_5Density()
	svc.AddCharacteristic(svc.PM2_5Density.Characteristic)
	return &svc
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
func NewAirQualitySensor(info accessory.Info, cur, min, max, steps float64) *AirQualitySensor {
	acc := AirQualitySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeAirPurifier)
	acc.AirQualitySensor = NewAirQualitySensorService()
	acc.AirQualitySensor.AirQuality.SetValue(0)
	acc.AirQualitySensor.AirQuality.SetMinValue(0)
	acc.AirQualitySensor.AirQuality.SetMaxValue(5)
	acc.AirQualitySensor.AirQuality.SetStepValue(1)
	acc.AirQualitySensor.PM2_5Density.SetValue(cur)
	acc.AirQualitySensor.PM2_5Density.SetMinValue(min)
	acc.AirQualitySensor.PM2_5Density.SetMaxValue(max)
	acc.AirQualitySensor.PM2_5Density.SetStepValue(steps)

	acc.AddService(acc.AirQualitySensor.Service)
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
			log.Println(acc)
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
				time.Sleep(time.Second * 10)
				continue
			}
			temp, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Println("bad value", value)
				time.Sleep(time.Second * 10)
				continue
			}
			log.Println("get value", value, info)
			acc.TempSensor.CurrentTemperature.SetValue(temp)
			time.Sleep(time.Second * 10)
		}
	case "HumiditySensor":
		acc := NewHumiditySensor(info, 5, 0, 200, 0.1)
		config := hc.Config{Pin: ac.Pin}
		t, err := hc.NewIPTransport(config, acc.Accessory)
		if err != nil {
			log.Println(acc)
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
				time.Sleep(time.Second * 10)
				continue
			}
			hum, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Println("bad value", value)
				time.Sleep(time.Second * 10)
				continue
			}
			log.Println("get value", value, info)
			acc.HumiditySensor.CurrentRelativeHumidity.SetValue(hum)
			time.Sleep(time.Second * 10)
		}
	case "AirQualitySensor":
		acc := NewAirQualitySensor(info, 0, 0, 2000, 1)
		config := hc.Config{Pin: ac.Pin}
		t, err := hc.NewIPTransport(config, acc.Accessory)
		if err != nil {
			log.Println(acc)
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
				time.Sleep(time.Second * 10)
				continue
			}
			cur, err := strconv.ParseFloat(value.(string), 64)
			if err != nil {
				log.Println("bad value", cur)
				time.Sleep(time.Second * 10)
				continue
			}
			log.Println("get value", value, info)
			acc.AirQualitySensor.PM2_5Density.SetValue(cur)
			if cur < 30 {
				acc.AirQualitySensor.AirQuality.SetValue(1)
			}
			if cur > 30 && cur < 60 {
				acc.AirQualitySensor.AirQuality.SetValue(2)
			}
			if cur > 60 && cur < 180 {
				acc.AirQualitySensor.AirQuality.SetValue(3)
			}
			if cur > 180 && cur < 280 {
				acc.AirQualitySensor.AirQuality.SetValue(4)
			}
			if cur > 280 {
				acc.AirQualitySensor.AirQuality.SetValue(5)
			}
			time.Sleep(time.Second * 10)
		}
	case "Switch":
		log.Println("test")
	}
}
