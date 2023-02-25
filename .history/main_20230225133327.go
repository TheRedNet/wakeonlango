package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	wol "github.com/sabhiram/go-wol/wol"
)

type Device struct {
	Name string `json:"name"`
	Mac  string `json:"mac"`
}

var editMode bool
var lastError string


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

func saveDevices(filename string, deviceMap map[string]string) error {
	// convert deviceMap to devices slice
	var devices []Device
	for name, mac := range deviceMap {
		devices = append(devices, Device{Name: name, Mac: mac})
	}
	// marshal devices slice to JSON data
	data, err := json.Marshal(devices)
	if err != nil {
		return err
	}
	// write data to file
	return ioutil.WriteFile(filename, data, 0644)
}

func sendMagicPacket(macAddr, bcastAddr string) (err error) {
	packet, err := wol.New(macAddr)
	if err != nil {
		return
	}
	payload, err := packet.Marshal()
	if err != nil {
		return
	}

	if bcastAddr == "" {
		bcastAddr = "255.255.255.255:9"
	}
	bcAddr, err := net.ResolveUDPAddr("udp", bcastAddr)
	if err != nil {
		return
	}

	conn, err := net.DialUDP("udp", nil, bcAddr)
	if err != nil {
		return
	}
	defer conn.Close()

	_, err = conn.Write(payload)
	return
}

func handleToggle(ctx *fiber.Ctx) error {
	editMode = !editMode
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		return ctx.Redirect("/")
	}
	var devices []Device
	for name, mac := range deviceMap {
		devices = append(devices, Device{name, mac})
	}
	return ctx.Redirect("/")
}

func handleEdit(ctx *fiber.Ctx) error {
	action := ctx.FormValue("action")
	name := ctx.FormValue("name")
	mac := ctx.FormValue("mac")
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		return ctx.Redirect("/")
	}
	switch action {
	case "add":
		deviceMap[name] = mac
	case "delete":
		delete(deviceMap, name)
	default:
		lastError = fmt.Sprintf("Invalid action: %s", action)
		return ctx.Redirect("/")
	}
	err = saveDevices(deviceMap, "devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to save devices.json: %v", err)
		return ctx.Redirect("/")
	}
	return ctx.Redirect("/")
}

func handleWake(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name")
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		return ctx.Redirect("/")
	}
	mac := deviceMap[name]
	if mac == "" {
		lastError = fmt.Sprintf("Unknown device: %s", name)
		return ctx.Redirect("/")
	}
	err = sendMagicPacket(mac, "255.255.255.255:9")
	if err != nil {
		lastError = fmt.Sprintf("Failed to send magic packet: %v", err)
		return ctx.Redirect("/")
	}
	return ctx.Redirect("/")
}
func handleIndex(ctx *fiber.Ctx) error {
	deviceMap, err := loadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		//return ctx.Redirect("/")
	} else {
	var devices []Device
	for name, mac := range deviceMap {
		devices = append(devices, Device{name, mac})
	}
	return ctx.Render("index", fiber.Map{
		"Devices":  devices,
		"EditMode": editMode,
		"LastError": lastError,
	})
}

func main() {
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Get("/", handleIndex)
	app.Post("/wake", handleWake)
	app.Post("/toggle", handleToggle)
	app.Listen(":8080")
}
