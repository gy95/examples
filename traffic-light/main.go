package main

import (
	"fmt"
	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"
	"log"
	"os"
	"time"
	"encoding/json"
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

	// twin get
	go func() {
		topic := DeviceETPrefix +  "default/" + deviceID + TwinETGetSuffix
		device := v1alpha2.Device{}
		for {
			fmt.Println("begin to update twin, topic is ", topic)
			twinGetBody, err := json.Marshal(device)
			if err != nil {
				log.Fatalf("Error:  %v", err)
			}
			token := Client.Publish(topic, 1, false, twinGetBody)
			if token.Wait() && token.Error() != nil {
				log.Fatalf("client.publish() Error in device twin update is %v", token.Error())
			}
			
			//fmt.Println("update deviceTwin %++v\n", string(twinUpdateBody))

			time.Sleep(time.Second * 60)
		}
	}()

	for {
		time.Sleep(time.Second * 2)
	}

	Client.Disconnect(250)
}
//8a24f255ed69848b2f8979ce2f2d795cbba2ca6e6e9e012b5fcd79a9203361a2.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTkwNTU2NzN9.R3O0omVUzNQdYoHAuSityFmjiVoCxUM1U3UE7jrnyiw