package main

import (
	"github.com/brutella/hc"
	"github.com/brutella/hc/accessory"
	"log"
)

func main() {
	info := accessory.Info{
		Name:         "DHT22",
		SerialNumber: "051AC-23AAM1",
		Manufacturer: "Apple",
		Model:        "AB",
	}
	acc := accessory.NewTemperatureSensor(info,5,-100,50,0.1)

	config := hc.Config{Pin: "00102003"}
	t, err := hc.NewIPTransport(config, acc.Accessory)
	if err != nil {
		log.Panic(err)
	}

	hc.OnTermination(func() {
		t.Stop()
	})

log.Println("set temp")
acc.TempSensor.CurrentTemperature.SetValue(10)
	t.Start()
}
