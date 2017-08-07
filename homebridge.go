package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/patrickmn/go-cache"
	"io/ioutil"
	"net/http"
	"os"
)

type HomekitBridge struct {
	ListenAddress string `json:"ListenAddress"`
	cache         *cache.Cache
	AccessoryList []Accessorys `json:"AccessoryList"`
}

func ReadConfig(file string) (*HomekitBridge, error) {
	configFile, err := os.Open(file)
	config, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	var s *HomekitBridge
	if err := json.Unmarshal(config, &s); err != nil {
		return nil, err
	}
	return s, nil
}

// /api/v1/accessory?name=%sysname%&task=%tskname%&valuename=%valname%&value=%value%
// demo.php?name=%sysname%&task=%tskname%&valuename=%valname%&value=%value%
func (hb *HomekitBridge) AccessoryUpdate(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=\"utf-8\"")
	accessoryName := c.Param("name")
	accessoryTask := c.Param("task")
	accessoryValueName := c.Param("valuename")
	accessoryValue := c.Param("value")
	hb.cache.Set(fmt.Sprintf("%s %s %s", accessoryName, accessoryTask, accessoryValueName), accessoryValue, cache.DefaultExpiration)
	fmt.Print(accessoryName, accessoryTask, accessoryValueName)
	c.JSON(http.StatusOK, gin.H{"status": "update info"})
}

func (hb *HomekitBridge) Tasks() {
	for _, ac := range hb.AccessoryList {
		go ac.Task()
	}
}
