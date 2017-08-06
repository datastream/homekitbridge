package main

import (
        "github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
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
	homekitBridge.db, err = sql.Open("postgres",homekitBridge.HomekitDatabaseURI)
        if err != nil {
                log.Fatal("database open:", err)
        }
        defer homekitBridge.db.Close()
	var id int
        err = homekitBridge.db.QueryRow("select count(*) from accessorys").Scan(&id)
        if err != nil && strings.Contains(err.Error(), "not exist") {
                err = databaseinit()
                if err != nil {
                        log.Fatal("database init create:", err)
                }
        }
        if err != nil {
                log.Fatal("database open:", err)
        }
	homekitBridge.cache = cache.New(5*time.Minute, 10*time.Minute)
	r := gin.Default()
	homekitAPI := r.Group("/api/v1")
	homekitAPI.GET("/accessory",homekitBridge.AccessoryUpdate)
	go homekitBridge.Tasks()
	r.Run(homekitBridge.ListenAddress)
}
