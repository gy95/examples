package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	parseFlag()

	loadConfigMap()

	InitCLient()

	if token := Client.Connect(); token.Wait() && token.Error() != nil {
		fmt.Println("connect info: ", token.Error())
		os.Exit(1)
	}

	//ChangeDeviceState("online")

	//GetTwin(CreateActualDeviceStatus(UNKNOW, UNKNOW, UNKNOW))


	// 先暂时注释掉这个，避免设备主动上报消息对调试产生影响
	go UpdateActualDeviceStatus()

	for {
		time.Sleep(time.Second * 2)
	}

	Client.Disconnect(250)
}
