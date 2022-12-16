package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/types"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype"
)

//var deviceID string = "light01"
var deviceID string
var mqtturl string
var modelName string

var actualColor string = ""

var CONFIG_MAP_PATH = "/opt/kubeedge/deviceProfile.json"

const (
	DeviceETPrefix            = "$hw/events/device/"
	TwinETUpdateSuffix        = "/twin/update"
	TwinETUpdateDetalSuffix   = "/twin/update/delta"
	DeviceETStateUpdateSuffix = "/state/update"
	TwinETCloudSyncSuffix     = "/twin/cloud_updated"
	TwinETGetResultSuffix     = "/twin/get/result"
	TwinETGetSuffix           = "/twin/get"
)

const (
	COLOR = "color"
)

var Client MQTT.Client
var onceClient sync.Once

func parseFlag() {
	flag.StringVar(&deviceID, "device", "light01", "device id name, default is light01")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://127.0.0.1:1883", "mqtt url default is tcp://127.0.0.1:1883")
	flag.StringVar(&modelName, "modelname", "light", "device model name , default is light")
	flag.Parse()
}

func loadConfigMap() error {

	readConfigMap := &types.DeviceProfile{}
	jsonFile, err := ioutil.ReadFile(CONFIG_MAP_PATH)
	if err != nil {
		log.Fatalf("readfile %v error %v\n", CONFIG_MAP_PATH, err)
		return err
	}
	err = json.Unmarshal(jsonFile, readConfigMap)
	if err != nil {
		log.Fatalf("unmarshal error %v", err)
		return err
	}

	for _, deviceModel := range readConfigMap.DeviceModels {
		if strings.ToUpper(deviceModel.Name) == strings.ToUpper(modelName) {
			for _, property := range deviceModel.Properties {
				name := strings.ToUpper(property.Name)
				if name == strings.ToUpper(COLOR) {
					if v, ok := property.DefaultValue.(string); !ok {
						log.Fatalf("get color error %v", property.DefaultValue)
						return errors.New(" Error in reading color from config map")
					} else {
						actualColor = string(v)
					}

				}

			}
		}
	}
	fmt.Printf("Finally get color from configmap: color %v \n", actualColor)

	SetOutput(actualColor)
	return nil
}

func InitCLient() MQTT.Client {
	fmt.Println("init client ...")
	onceClient.Do(func() {
		opts := MQTT.NewClientOptions().AddBroker(mqtturl).SetClientID("test").SetCleanSession(false)
		opts = opts.SetKeepAlive(10)
		opts = opts.SetOnConnectHandler(func(c MQTT.Client) {
			topic := DeviceETPrefix + deviceID + TwinETUpdateDetalSuffix
			if token := c.Subscribe(topic, 0, OperateUpdateDetalSub); token.Wait() && token.Error() != nil {
				fmt.Println("subscribe: ", token.Error())
				os.Exit(1)
			}
		})
		Client = MQTT.NewClient(opts)
	})
	return Client
}

func OperateUpdateDetalSub(c MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Receive msg topic %s %v\n\n", msg.Topic(), string(msg.Payload()))
	current := &dttype.DeviceTwinUpdate{}
	if err := json.Unmarshal(msg.Payload(), current); err != nil {
		fmt.Printf("unmarshl receive msg DeviceTwinUpdate{} to error %v\n", err)
		return
	}
	fmt.Println("get message is ", *current)
	value := *(current.Twin[COLOR].Expected.Value)

	if actualColor != value {
		if err := SetOutput(value); err != nil {
			fmt.Printf("set light color to %v error %v", value, err)
		}
	}
	actualColor = value
}

func CreateActualDeviceStatus(color string) dttype.DeviceTwinUpdate {
	act := dttype.DeviceTwinUpdate{}
	actualMap := map[string]*dttype.MsgTwin{
		COLOR: {
			Actual:   &dttype.TwinValue{Value: &color},
			Metadata: &dttype.TypeMetadata{Type: "Updated"},
		},
	}
	act.Twin = actualMap
	return act
}

func UpdateActualDeviceStatus() {
	deviceTwinUpdate := DeviceETPrefix + deviceID + TwinETUpdateSuffix
	for {
		act := CreateActualDeviceStatus(actualColor)

		twinUpdateBody, err := json.Marshal(act)
		if err != nil {
			log.Fatalf("Error:  %v", err)
		}
		token := Client.Publish(deviceTwinUpdate, 1, false, twinUpdateBody)
		if token.Wait() && token.Error() != nil {
			log.Fatalf("client.publish() Error in device twin update is %v", token.Error())
		}

		time.Sleep(time.Second * 3)
	}

}

//DeviceStateUpdate is the structure used in updating the device state
type DeviceStateUpdate struct {
	State string `json:"state,omitempty"`
}
