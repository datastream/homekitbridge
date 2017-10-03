package main

import (
	"flag"
	"github.com/patrickmn/go-cache"
	"github.com/surgemq/surgemq/service"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	homekitBridge.cache = cache.New(5*time.Minute, 10*time.Minute)
	// Create a new server
	svr := &service.Server{
		KeepAlive:        600,           // seconds
		ConnectTimeout:   5,             // seconds
		SessionsProvider: "mem",         // keeps sessions in memory
		Authenticator:    "mockSuccess", // always succeed
		TopicsProvider:   "mem",         // keeps topic subscriptions in memory
	}
	go svr.ListenAndServe(homekitBridge.ListenAddress)
	go homekitBridge.Tasks()
	termchan := make(chan os.Signal, 1)
	signal.Notify(termchan, syscall.SIGINT, syscall.SIGTERM)
	<-termchan
}
