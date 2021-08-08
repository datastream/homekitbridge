package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	confFile = flag.String("c", "homekit.json", "homekit service config file")
)

var homekitBridge *HomekitBridge

func main() {
	flag.Parse()
	var err error
	homekitBridge, err = ReadConfig(*confFile)
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(homekitBridge.MetricListenAddress, nil)
	go homekitBridge.Tasks()
	termchan := make(chan os.Signal, 1)
	signal.Notify(termchan, syscall.SIGINT, syscall.SIGTERM)
	<-termchan
}
