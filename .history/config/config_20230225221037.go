package config

import (
	"encoding/json"
	"io/ioutil"
)

type Device struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
}

type Config struct {
	Devices       []Device `json:"devices"`
	IsAuthEnabled bool     `json:"isAuthEnabled"`
	Username      string   `json:"username"`
	Password      string   `json:"password"`
}

func LoadDevices(filename string) (map[string]string, error) {
	config, err := loadConfig(filename)
	var devices []Device
	devices = config.Devices
	deviceMap := make(map[string]string)
	for _, d := range devices {
		deviceMap[d.Name] = d.Mac
	}
	return deviceMap, err
}

func SaveDevices(filename string, deviceMap map[string]string) error {
	// convert deviceMap to devices slice
	var devices []Device
	for name, mac := range deviceMap {
		devices = append(devices, Device{Name: name, Mac: mac})
	}
	var config Config
	config, err := loadConfig(filename)
	config.Devices = devices
	err = saveConfig(filename, config)
	return err
}

func LoadConfig(filename string) (Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return Config{}, err
	}
	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}
func SaveConfig(filename string, config Config) error {
	// marshal config to JSON data
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	// write data to file
	return ioutil.WriteFile(filename, data, 0644)
}
