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
	Topic         string `json:"topic"`
	Name          string `json:"name"`
	SerialNumber  string `json:"serialNumber"`
	Manufacturer  string `json:"manufacturer"`
	Model         string `json:"model"`
	Pin           string `json:"pin"`
	AccessoryType string `json:"accessoryType"`
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
	acc.Accessory = accessory.New(info, accessory.TypeHumidifer)
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
			log.Println(ac.Topic, value)
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
			log.Println(ac.Topic, value)
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
			log.Println(ac.Topic, value)
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
			log.Println(ac.Topic, value)
			acc.CarbonDioxideSensor.CarbonDioxideLevel.SetValue(value)
			if acc.CarbonDioxideSensor.CarbonDioxidePeakLevel.GetValue() < value {
				acc.CarbonDioxideSensor.CarbonDioxidePeakLevel.SetValue(value)
			}
			if value > 1000 {
				acc.CarbonDioxideSensor.CarbonDioxideDetected.SetValue(1)
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
	opts.SetOnConnectHandler(ac.onConnect)
	opts.SetConnectionLostHandler(ac.onLost)
	ac.client = mqtt.NewClient(opts)
	if token := ac.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	fmt.Println("finish readmqtt")
	return nil
}

// openHAB MQTT
// /%sysname%/%tskname%/%valname%
func (ac *Accessorys) AccessoryUpdate(client mqtt.Client, msg mqtt.Message) {
	payload := msg.Payload()
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

func (ac *Accessorys) onLost(client mqtt.Client, err error) {
	fmt.Println(err, ac.Topic)
}

func (ac *Accessorys) onConnect(client mqtt.Client) {
	if token := ac.client.Subscribe(ac.Topic, byte(0), nil); token.Wait() && token.Error() != nil {
		fmt.Println(token.Error())
	}
}
