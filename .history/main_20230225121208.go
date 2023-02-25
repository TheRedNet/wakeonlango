package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	wol "github.com/sabhiram/go-wol/wol"
)

type Device struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
}

func loadDevices(filename string) (map[string]string, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var devices []Device
	err = json.Unmarshal(data, &devices)
	if err != nil {
		return nil, err
	}
	deviceMap := make(map[string]string)
	for _, d := range devices {
		deviceMap[d.Name] = d.Mac
	}
	return deviceMap, nil
}

func sendMagicPacket(mac string) error {
	return wol.SendMagicPacket(mac, "255.255.255.255:9")
}

func handleRequest(ctx *fiber.Ctx) error {
	name := ctx.Query("device")
	if name == "" {
		return ctx.SendString("Missing device parameter")
	}
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		return ctx.SendString(fmt.Sprintf("Failed to load devices.json: %v", err))
	}
	mac := deviceMap[name]
	if mac == "" {
		return ctx.SendString(fmt.Sprintf("Unknown device: %s", name))
	}
	err = sendMagicPacket(mac)
	if err != nil {
		return ctx.SendString(fmt.Sprintf("Failed to send magic packet: %v", err))
	}
	return ctx.SendString("OK")
}

func handleIndex(ctx *fiber.Ctx) error {
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		return ctx.SendString(fmt.Sprintf("Failed to load devices.json: %v", err))
	}
	var devices []Device
	for name, mac := range deviceMap {
		devices = append(devices, Device{name, mac})
	}
	return ctx.Render("index", fiber.Map{
		"Devices": devices,
	})
}

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/", handleIndex)
	app.Get("/wake", handleRequest)
	app.Listen(":8080")
}
