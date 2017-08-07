package main

import (
	"flag"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
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
	r := gin.Default()
	homekitAPI := r.Group("/api/v1")
	homekitAPI.GET("/accessory", homekitBridge.AccessoryUpdate)
	go homekitBridge.Tasks()
	go r.Run(homekitBridge.ListenAddress)
	termchan := make(chan os.Signal, 1)
	signal.Notify(termchan, syscall.SIGINT, syscall.SIGTERM)
	<-termchan
}
