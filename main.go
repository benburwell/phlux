package main

import (
	"log"
	"os"
	"time"

	hue "github.com/benburwell/gohue"
)

const username = "phlux"

func main() {
	var config PhluxConfig
	config.Read()
	log.Println("Config:", config)

	bridges, err := hue.FindBridges()
	if err != nil {
		log.Fatalf("Error finding bridges: %s\n", err.Error())
	}
	log.Printf("Found %d bridge(s)\n", len(bridges))
	desiredColorTemp := getDesiredColorTemperature(time.Now(), config.Latitude, config.Longitude)
	for _, bridge := range bridges {
		log.Printf("Bridge: %s\n", bridge.IPAddress)
		updateBridge(bridge, desiredColorTemp)
	}
}

func updateBridge(bridge hue.Bridge, ct ColorTemperature) {
	//username, err := bridge.CreateUser(username)
	//if err != nil {
	//	panic("Could not create user on bridge")
	//}
	//fmt.Printf("Made user %s for bridge %s\n", username, bridge.IPAddress)
	err := bridge.Login(os.Getenv("HUE_LOGIN"))
	if err != nil {
		log.Fatalf("Could not log in to bridge: %s\n", err.Error())
	}
	log.Println("Logged in to bridge")

	config, err := bridge.GetConfig()
	if err != nil {
		log.Fatalf("Could not get bridge info: %s\n", err.Error())
	}
	log.Printf("Bridge name: %s\n", config.Name)
	log.Printf("Bridge ID: %s\n", config.BridgeID)

	lights, err := bridge.GetAllLights()
	if err != nil {
		log.Fatalf("Error getting lights: %s\n", err.Error())
	}
	log.Printf("Found %d lights\n", len(lights))
	for _, light := range lights {
		updateLight(light, ct)
	}
}

func updateLight(light hue.Light, ct ColorTemperature) {
	log.Printf("Light %d: %s (%s)\n", light.Index, light.Name, light.Type)
	if supportsColorTemp(light) {
		log.Printf("  CT range: %d-%d\n", light.Capabilities.Control.CT.Min, light.Capabilities.Control.CT.Max)
		newCt := ct.TranslateForLight(light)
		log.Printf("  Setting CT to %d\n", newCt)
		light.SetState(hue.LightState{
			On: light.State.On,
			CT: newCt,
		})
	}
}
