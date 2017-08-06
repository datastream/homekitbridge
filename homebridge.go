package main

import (
        "github.com/gin-gonic/gin"
	"time"
)

type HomekitBridge struct {
	HomekitDatabaseURI string `json:"HomekitDatabaseURI"`
	ListenAddress string `json:"ListenAddress"`
	db *sql.DB
	cache *cache.Cache
	taskList map[int]int
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
// demo.php?name=%sysname%&task=%tskname%&valuename=%valname%&value=%value%
func (hc *HomekitBridge)AccessoryUpdate(c *gin.Context) {
	c.Header("Content-Type", "application/json; charset=\"utf-8\"")
	accessoryName := c.Param("name")
	accessoryTask := c.Param("task")
	accessoryValueName := c.Param("valuename")
	accessoryValue := c.Param("value")
	hc.Set(fmt.Sprintf("%s %s %s", accessoryName, accessoryTask, accessoryValueName),accessoryValue,cache.DefaultExpiration)
	fmt.Print(accessoryName, accessoryTask, accessoryValueName)
	c.JSON(http.StatusOK, gin.H{"status": "update info"})
}

func (hc *HomekitBridge) Tasks() {
	for {
		tasks := make(map[int]int)
		accessorys , _:= FindAllAccessorys()
		for _, accessory := range accessorys {
			tasks[accessory.ID] = accessory.ID
			if _, ok := hc.taskList[accessory.ID]; ok {
				continue
			}
			hc.taskList[accessory.ID] = accessory.ID
			go accessory.Task()
		}
		for k := range tasks {
			if _, ok := hc.taskList[k]; !ok {
				delete(hc.taskList, k)
			}
		}
		time.Sleep(30*time.Second)
	}
}
