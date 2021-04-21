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
//d30f363172f0b03011e769c8eb8e21396a6765ce53066bce623158590b557eac.eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTkwNTczMjN9.qkQ_oVsVO2IxNKHnKQh5XThc11Rly6KMC6OJ_3rG2to