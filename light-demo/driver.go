package main

import (
	"encoding/json"
	"os"
)

type Result struct {
	Device Device `json:"device,omitempty"`
}

type Device struct {
	Name  string `json:"name,omitempty"`
	Color string `json:"color,omitempty"`
}

// 格式化输出json结果到文件
func SetOutput(color string) error {
	result := &Result{
		Device: Device{
			Name:  deviceID,
			Color: color,
		},
	}

	// 以json格式输出到本地文件
	data, err := json.MarshalIndent(result, "", "	")
	if err != nil {
		return err
	}

	err = os.WriteFile("/result/result.json", data, 0644)
	if err != nil {
		return err
	}

	return nil
}
