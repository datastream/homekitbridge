package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type Accessorys struct {
	hb             *HomekitBridge
	dataChannel    chan float64
	exitChan       chan int
	MetricEndpoint string `json:"metricEndpoint"`
	Metric         string `json:"metric"`
	Name           string `json:"name"`
	SerialNumber   string `json:"serialNumber"`
	Manufacturer   string `json:"manufacturer"`
	Model          string `json:"model"`
	Pin            string `json:"pin"`
	AccessoryType  string `json:"accessoryType"`
}

type CarbonDioxideSensorService struct {
	*service.Service
	CarbonDioxideDetected  *characteristic.CarbonDioxideDetected
	CarbonDioxideLevel     *characteristic.CarbonDioxideLevel
	CarbonDioxidePeakLevel *characteristic.CarbonDioxidePeakLevel
}

func NewCarbonDioxideSensorService() *CarbonDioxideSensorService {
	svc := CarbonDioxideSensorService{}
	svc.Service = service.New(service.TypeCarbonDioxideSensor)
	svc.CarbonDioxideDetected = characteristic.NewCarbonDioxideDetected()
	svc.AddCharacteristic(svc.CarbonDioxideDetected.Characteristic)
	svc.CarbonDioxideLevel = characteristic.NewCarbonDioxideLevel()
	svc.AddCharacteristic(svc.CarbonDioxideLevel.Characteristic)
	svc.CarbonDioxidePeakLevel = characteristic.NewCarbonDioxidePeakLevel()
	svc.AddCharacteristic(svc.CarbonDioxidePeakLevel.Characteristic)
	return &svc
}

type CarbonDioxideSensor struct {
	*accessory.Accessory
	CarbonDioxideSensor *CarbonDioxideSensorService
}

func NewCarbonDioxideSensor(info accessory.Info) *CarbonDioxideSensor {
	acc := CarbonDioxideSensor{}
	acc.Accessory = accessory.New(info, accessory.TypeAirPurifier)
	acc.CarbonDioxideSensor = NewCarbonDioxideSensorService()
	acc.AddService(acc.CarbonDioxideSensor.Service)
	return &acc
}

type HumiditySensor struct {
	*accessory.Accessory

	HumiditySensor *service.HumiditySensor
}

func NewHumiditySensor(info accessory.Info) *HumiditySensor {
	acc := HumiditySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeHumidifier)
	acc.HumiditySensor = service.NewHumiditySensor()

	acc.AddService(acc.HumiditySensor.Service)

	return &acc
}

type AirQualitySensorService struct {
	*service.Service
	AirQuality            *characteristic.AirQuality
	AirParticulateDensity *characteristic.AirParticulateDensity
	AirParticulateSize    *characteristic.AirParticulateSize
}

func NewAirQualitySensorService() *AirQualitySensorService {
	svc := AirQualitySensorService{}
	svc.Service = service.New(service.TypeAirQualitySensor)
	svc.AirQuality = characteristic.NewAirQuality()
	svc.AddCharacteristic(svc.AirQuality.Characteristic)
	svc.AirParticulateDensity = characteristic.NewAirParticulateDensity()
	svc.AddCharacteristic(svc.AirParticulateDensity.Characteristic)
	svc.AirParticulateSize = characteristic.NewAirParticulateSize()
	svc.AddCharacteristic(svc.AirParticulateSize.Characteristic)
	return &svc
}

type AirQualitySensor struct {
	*accessory.Accessory

	AirQualitySensor *AirQualitySensorService
}

func NewAirQualitySensor(info accessory.Info) *AirQualitySensor {
	acc := AirQualitySensor{}
	acc.Accessory = accessory.New(info, accessory.TypeAirPurifier)
	acc.AirQualitySensor = NewAirQualitySensorService()
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
	go ac.AccessoryUpdate()
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
		for value := range ac.dataChannel {
			log.Println(ac.Metric, value)
			acc.TempSensor.CurrentTemperature.SetValue(value)
		}
	case "HumiditySensor":
		acc := NewHumiditySensor(info)
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
		for value := range ac.dataChannel {
			log.Println(ac.Metric, value)
			acc.HumiditySensor.CurrentRelativeHumidity.SetValue(value)
		}
	case "AirQualitySensor":
		acc := NewAirQualitySensor(info)
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
		for value := range ac.dataChannel {
			log.Println(ac.Metric, value)
			acc.AirQualitySensor.AirParticulateDensity.SetValue(value)
			if value <= 50 {
				acc.AirQualitySensor.AirQuality.SetValue(1)
			}
			if value > 50 && value <= 100 {
				acc.AirQualitySensor.AirQuality.SetValue(2)
			}
			if value > 100 && value <= 150 {
				acc.AirQualitySensor.AirQuality.SetValue(3)
			}
			if value > 150 && value <= 200 {
				acc.AirQualitySensor.AirQuality.SetValue(4)
			}
			if value > 200 {
				acc.AirQualitySensor.AirQuality.SetValue(5)
			}
		}
	case "CarbonDioxideSensor":
		acc := NewCarbonDioxideSensor(info)
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
		for value := range ac.dataChannel {
			log.Println(ac.Metric, value)
			acc.CarbonDioxideSensor.CarbonDioxideLevel.SetValue(value)
			if acc.CarbonDioxideSensor.CarbonDioxidePeakLevel.GetValue() < value {
				acc.CarbonDioxideSensor.CarbonDioxidePeakLevel.SetValue(value)
			}
			if value > 1200 {
				acc.CarbonDioxideSensor.CarbonDioxideDetected.SetValue(1)
			} else {
				acc.CarbonDioxideSensor.CarbonDioxideDetected.SetValue(0)
			}
		}
	case "Switch":
		log.Println("test")
	}
}

func (ac *Accessorys) AccessoryUpdate() {
	client, err := api.NewClient(api.Config{
		Address: ac.MetricEndpoint,
	})
	if err != nil {
		log.Fatal(err.Error())
	}
	v1api := v1.NewAPI(client)
	ticker := time.Tick(time.Minute)
	for {
		select {
		case <-ticker:
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			current := time.Now()
			r := v1.Range{
				Start: current.Add(-time.Minute * 2),
				End:   current,
				Step:  time.Minute,
			}
			result, warnings, err := v1api.QueryRange(ctx, ac.Metric, r)
			cancel()
			if err != nil {
				fmt.Printf("Error querying Prometheus: %v\n", err)
			}
			if len(warnings) > 0 {
				fmt.Printf("Warnings: %v\n", warnings)
			}
			log.Println(result.String())
			items := strings.Split(result.String(), "\n")
			l := len(items)
			var value float64
			if l > 2 {
				data := items[l-1]
				values := strings.Split(data, " ")
				log.Println(data, values)
				if len(values) < 2 {
					break
				}
				value, err = strconv.ParseFloat(values[0], 64)
				if err != nil {
					log.Println("convert to float64", err)
					break
				}
			}
			log.Println(value)
			ac.dataChannel <- value
		case <-ac.exitChan:
		}
	}
}
