package main

import (
	"fmt"
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"github.com/brutella/hc/characteristic"
	"github.com/brutella/hc/service"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"strconv"
)

type Accessorys struct {
	hb            *HomekitBridge
	client        mqtt.Client
	dataChannel   chan float64
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
	AirQuality            *characteristic.AirQuality
	AirParticulateDensity *characteristic.AirParticulateDensity
	AirParticulateSize    *characteristic.AirParticulateSize
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
	svc.AirParticulateDensity = characteristic.NewAirParticulateDensity()
	svc.AddCharacteristic(svc.AirParticulateDensity.Characteristic)
	svc.AirParticulateSize = characteristic.NewAirParticulateSize()
	svc.AddCharacteristic(svc.AirParticulateSize.Characteristic)
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
	go ac.ReadMQTT()
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
			log.Println(ac.Key, value)
			acc.TempSensor.CurrentTemperature.SetValue(value)
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
		for value := range ac.dataChannel {
			log.Println(ac.Key, value)
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
			log.Println(ac.Key, value)
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
	case "Switch":
		log.Println("test")
	}
}

func (ac *Accessorys) ReadMQTT() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(ac.hb.ListenAddress)
	opts.SetClientID(fmt.Sprintf("homebirdge%s", ac.Name))
	opts.SetUsername(ac.hb.UserName)
	opts.SetPassword(ac.hb.Password)
	opts.SetCleanSession(true)
	opts.SetDefaultPublishHandler(ac.AccessoryUpdate)
	ac.client = mqtt.NewClient(opts)
	if token := ac.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	if token := ac.client.Subscribe(ac.Key, byte(0), nil); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	fmt.Println("finish readmqtt")
	return nil
}

// openHAB MQTT
// /%sysname%/%tskname%/%valname%
func (ac *Accessorys) AccessoryUpdate(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
	topic := msg.Topic()
	if string(payload) == "Connected" {
		return
	}
	value, err := strconv.ParseFloat(string(payload), 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	ac.dataChannel <- value
}
