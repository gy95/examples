package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/kubeedge/kubeedge/cloud/pkg/apis/devices/v1alpha2"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/kubeedge/kubeedge/cloud/pkg/devicecontroller/types"

	MQTT "github.com/eclipse/paho.mqtt.golang"

	"github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype"
)

//var deviceID string = "traffic-light-instance-01"
var deviceID string
var mqtturl string
var modelName string

var red_wpi_num int64 = 25
var yellow_wpi_num int64 = 24
var green_wpi_num int64 = 23

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
	RED_STATE    = "red"
	YELLOW_STATE = "yellow"
	GREEN_STATE  = "green"
)

const (
	RedpinNumberConfig    = "red-pin-number"
	YellowpinNumberConfig = "yellow-pin-number"
	GreenpinNumberConfig  = "green-pin-number"
)

var Client MQTT.Client
var onceClient sync.Once

func parseFlag() {
	flag.StringVar(&deviceID, "device", "traffic-light-instance-pi1", "device id name, default is traffic-light-instance-pi1 ")
	flag.StringVar(&mqtturl, "mqtturl", "tcp://127.0.0.1:1883", "mqtt url default is tcp://127.0.0.1:1883")
	flag.StringVar(&modelName, "modelname", "traffic-light", "device model name , default is traffic-light")
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
				if name == strings.ToUpper(RedpinNumberConfig) {
					if num, ok := property.DefaultValue.(float64); !ok {
						log.Fatalf("get red pin number error %v", property.DefaultValue)
						return errors.New(" Error in reading red pin number from config map")
					} else {
						red_wpi_num = int64(num)
					}

				}
				if name == strings.ToUpper(YellowpinNumberConfig) {
					if num, ok := property.DefaultValue.(float64); !ok {
						log.Fatalf("get yellow pin number error ")
						return errors.New(" Error in reading yellow pin number from config map")
					} else {
						yellow_wpi_num = int64(num)
					}
				}
				if name == strings.ToUpper(GreenpinNumberConfig) {
					if num, ok := property.DefaultValue.(float64); !ok {
						log.Fatalf("get green pin number error ")
						return errors.New(" Error in reading green pin number from config map")
					} else {
						green_wpi_num = int64(num)
					}
				}
			}
		}
	}
	fmt.Printf("Finally get wpi pin number from configmap: red %d yellow %d green %d\n",
		red_wpi_num, yellow_wpi_num, green_wpi_num)

	SetOutput(red_wpi_num)
	SetOutput(yellow_wpi_num)
	SetOutput(green_wpi_num)
	return nil
}

func InitCLient() MQTT.Client {
	fmt.Println("init client ...")
	onceClient.Do(func() {
		opts := MQTT.NewClientOptions().AddBroker(mqtturl).SetClientID("zhangjie-test").SetCleanSession(false)
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

// this structure is used not only for cloud, but also for edge, so put here temporarily
type DeviceTransmitMsg struct {
	//BaseMessage BaseMessage
	// TODO: if need slice, i find only one device will be added or removed per one time
	// todo: update/add device all use this structure, considering add a method
	DeviceAdded v1alpha2.Device
	DeviceRemoved v1alpha2.Device
	// todo: consider move DeviceUpdated to other structure
	DeviceUpdated v1alpha2.Device
}

func OperateUpdateDetalSub(c MQTT.Client, msg MQTT.Message) {
	fmt.Printf("Receive msg topic %s %v\n\n", msg.Topic(), string(msg.Payload()))
	//current := &v1alpha2.DeviceStatus{}
	current := &DeviceTransmitMsg{}
	if err := json.Unmarshal(msg.Payload(), current); err != nil {
		fmt.Printf("unmarshl receive msg DeviceTwinUpdate{} to error %v\n", err)
		return
	}
	fmt.Printf("hhhhhhh")
	fmt.Printf("revceive msg is %v", current)
	twins := []v1alpha2.Twin{}
	if !reflect.DeepEqual(current.DeviceAdded, v1alpha2.Device{}) {
		twins = current.DeviceAdded.Status.Twins
	} else if !reflect.DeepEqual(current.DeviceUpdated, v1alpha2.Device{}) {
		twins = current.DeviceUpdated.Status.Twins
	} else if !reflect.DeepEqual(current.DeviceUpdated, v1alpha2.Device{}) {
		twins = current.DeviceRemoved.Status.Twins
	}
	fmt.Printf("after parsed, twins become %v", twins)
	for _, twin := range twins {
		if twin.PropertyName == RED_STATE {
			value := twin.Desired.Value
			fmt.Printf("red value is %s", value)
			if LedState(red_wpi_num) != value {
				if err := Set(red_wpi_num, value); err != nil {
					fmt.Printf("Set Red light to %v error %v", value, err)
				}
			}
		} else if twin.PropertyName == YELLOW_STATE {
			value := twin.Desired.Value
			fmt.Printf("yellow value is %s", value)
			if LedState(yellow_wpi_num) != value {
				if err := Set(yellow_wpi_num, value); err != nil {
					fmt.Printf("Set Yellow light to %v error %v", value, err)
				}
			}
		} else if twin.PropertyName == GREEN_STATE {
			value := twin.Desired.Value
			fmt.Printf("green value is %s", value)
			if LedState(green_wpi_num) != value {
				if err := Set(green_wpi_num, value); err != nil {
					fmt.Printf("Set Green light to %v error %v", value, err)
				}
			}
		}
	}
}

func CreateActualDeviceStatus(actred, actyellow, actgreen string) dttype.DeviceTwinUpdate {
	act := dttype.DeviceTwinUpdate{}
	actualMap := map[string]*dttype.MsgTwin{
		RED_STATE: {
			Actual:   &dttype.TwinValue{Value: &actred},
			Metadata: &dttype.TypeMetadata{Type: "Updated"}},
		YELLOW_STATE: {
			Actual:   &dttype.TwinValue{Value: &actyellow},
			Metadata: &dttype.TypeMetadata{Type: "Updated"}},
		GREEN_STATE: {
			Actual:   &dttype.TwinValue{Value: &actgreen},
			Metadata: &dttype.TypeMetadata{Type: "Updated"}},
	}
	act.Twin = actualMap
	return act
}

func LedState(number int64) string {
	s, err := State(number)
	if err != nil {
		log.Fatalf("get Led %d State  error %v", number, err)
	}
	switch s[0] {
	case '0':
		return "OFF"
	case '1':
		return "ON"
	}
	return UNKNOW
}

func UpdateActualDeviceStatus() {
	//r .y. g

	deviceTwinUpdate := DeviceETPrefix + deviceID + TwinETUpdateSuffix
	for {
		act := CreateActualDeviceStatus(LedState(red_wpi_num), LedState(yellow_wpi_num), LedState(green_wpi_num))

		fmt.Printf("begin to update twin")
		//twinUpdateBody, err := json.MarshalIndent(act, "", "	")
		twinUpdateBody, err := json.Marshal(act)
		if err != nil {
			log.Fatalf("Error:  %v", err)
		}
		token := Client.Publish(deviceTwinUpdate, 1, false, twinUpdateBody)
		if token.Wait() && token.Error() != nil {
			log.Fatalf("client.publish() Error in device twin update is %v", token.Error())
		}

		//fmt.Printf("update deviceTwin %++v\n", string(twinUpdateBody))

		time.Sleep(time.Second * 3000)
	}

}

//DeviceStateUpdate is the structure used in updating the device state
type DeviceStateUpdate struct {
	State string `json:"state,omitempty"`
}

/*
func ChangeDeviceState(state string) {
	fmt.Println("Changing the state of the device to online")
	var deviceStateUpdateMessage DeviceStateUpdate
	deviceStateUpdateMessage.State = state
	stateUpdateBody, err := json.Marshal(deviceStateUpdateMessage)
	if err != nil {
		log.Fatalf("Error:   %v", err)
	}
	deviceStatusUpdate := DeviceETPrefix + deviceID + DeviceETStateUpdateSuffix
	token := Client.Publish(deviceStatusUpdate, 0, false, stateUpdateBody)
	if token.Wait() && token.Error() != nil {
		log.Fatalf("client.publish() Error in device state update  is  %v", token.Error())
	}
}
 */

//getTwin function is used to get the device twin details from the edge
/*
func GetTwin(updateMessage dttype.DeviceTwinUpdate) {
	getTwin := DeviceETPrefix + deviceID + TwinETGetSuffix
	twinUpdateBody, err := json.Marshal(updateMessage)
	if err != nil {
		log.Fatalf("Error %v", err)
	}
	token := Client.Publish(getTwin, 0, false, twinUpdateBody)
	if token.Wait() && token.Error() != nil {
		log.Fatalf("client.publish() Error in device twin get  is ", token.Error())
	}
}
 */
