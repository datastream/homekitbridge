# HomekitBridge

build for rasbian
`GOOS=linux GOARCH=arm GOARM=6 go build`

# Support
1. support easyesp 2.0's `Generic HTTP`
2. support mqtt (todo)

# API
`/api/v1/accessory?name=%sysname%&task=%tskname%&valuename=%valname%&value=%value%`

`name=sanctum&taskname=DHT22&valuename=Humidity` -> json config file's `"key":"sanctum DHT22 Humidity"`
