package main

import (
	"fmt"
	"net"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
	wol "github.com/sabhiram/go-wol/wol"
	"therednet.de/wakeonlan/config"
)

var editMode bool
var lastError string

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
	deviceMap, err := config.LoadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		return ctx.Redirect("/")
	}
	var devices []config.Device
	for name, mac := range deviceMap {
		devices = append(devices, config.Device{Name: name, Mac: mac})
	}
	return ctx.Redirect("/")
}

func handleEdit(ctx *fiber.Ctx) error {
	action := ctx.FormValue("action")
	name := ctx.FormValue("name")
	mac := ctx.FormValue("mac")
	deviceMap, err := config.LoadDevices("devices.json")
	if err != nil {
		deviceMap = make(map[string]string)
		lastError = fmt.Sprintf("Failed to load devices.json: %v. Created a new one", err)
		//return ctx.Redirect("/")
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
	err = config.SaveDevices("devices.json", deviceMap)
	if err != nil {
		lastError = fmt.Sprintf("Failed to save devices.json: %v", err)
		return ctx.Redirect("/")
	}
	return ctx.Redirect("/")
}

func handleWake(ctx *fiber.Ctx) error {
	name := ctx.FormValue("name")
	deviceMap, err := config.LoadDevices("devices.json")
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
	deviceMap, err := config.LoadDevices("devices.json")
	if err != nil {
		lastError = fmt.Sprintf("Failed to load devices.json: %v", err)
		//return ctx.Redirect("/")
	}
	var devices []config.Device
	for name, mac := range deviceMap {
		devices = append(devices, config.Device{Name: name, Mac: mac})
	}
	displayedError := lastError
	lastError = ""
	hasErrored := displayedError != ""
	editModebool := editMode
	if hasErrored {
		println(displayedError)
	}
	return ctx.Render("index", fiber.Map{
		"Devices":    devices,
		"EditMode":   editModebool,
		"LastError":  displayedError,
		"HasErrored": hasErrored,
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
	app.Post("/edit", handleEdit)
	app.Listen("127.0.0.1: