package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	hue "github.com/benburwell/gohue"
)

const USERNAME = "phlux"

func main() {
	var config PhluxConfig
	config.Read()

	flag.Float64Var(&config.Latitude, "latitude", config.Latitude, "Latitude (used to calculate sunrise and sunset)")
	flag.Float64Var(&config.Longitude, "longitude", config.Longitude, "Longitude (used to calculate sunrise and sunset)")
	flag.Int64Var(&config.Interval, "interval", config.Interval, "Interval (in seconds) at which to run the update if --forever is set")
	flag.StringVar(&config.TransitionTime, "transitionTime", config.TransitionTime, "Transition time for adjusting the color temperature")
	forever := flag.Bool("forever", false, "Run the update every [interval], use Ctrl-C to cancel")
	flag.Parse()

	log.Println("Config:", config)

	updateColorTemps(&config)
	if *forever {
		ticker := time.NewTicker(time.Duration(config.Interval) * time.Second)
		for {
			<-ticker.C
			updateColorTemps(&config)
		}
	}
}

func updateColorTemps(config *PhluxConfig) {
	bridges, err := hue.FindBridges()
	if err != nil {
		log.Fatalf("Error finding bridges: %s\n", err.Error())
	}
	log.Printf("Found %d bridge(s)\n", len(bridges))
	desiredColorTemp := getDesiredColorTemperature(time.Now(), config.Latitude, config.Longitude)
	for _, bridge := range bridges {
		log.Printf("Bridge: %s\n", bridge.IPAddress)
		updateBridge(&bridge, desiredColorTemp, config)
	}
}

// In case we don't know the bridge's serial number for some reason, we won't
// be able to look up the token, nor will we be able to save it. There is still
// a chance we will be able to successfully log in though: if the link button
// has been pressed, we should be able to create ourselves a temporary token.
func authenticateOnce(bridge *hue.Bridge) error {
	token, err := bridge.CreateUser(USERNAME)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not create temporary token: %s", err.Error()))
	}
	log.Printf("Made token %s for bridge %s\n", token, bridge.IPAddress)
	err = bridge.Login(token)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to log in with temporary token: %s", err.Error()))
	}
	log.Printf("Logged in to bridge %s\n", bridge.IPAddress)
	return nil
}

func createToken(bridge *hue.Bridge, config *PhluxConfig) error {
	if err := authenticateOnce(bridge); err != nil {
		return err
	}
	config.SetBridgeToken(bridge.Info.Device.SerialNumber, bridge.Username)
	config.Save()
	return nil
}

// Attempt to authenticate to the bridge using a variety of techniques,
// including looking up a saved token for the bridge's serial number and
// attempting to generate a new token for the bridge assuming the link button
// has been pressed.
func authenticate(bridge *hue.Bridge, config *PhluxConfig) error {
	// get bridge info, which contains serial number
	err := bridge.GetInfo()
	if err != nil {
		return authenticateOnce(bridge)
	}
	token, err := config.GetBridgeToken(bridge.Info.Device.SerialNumber)
	if err != nil {
		return createToken(bridge, config)
	}
	err = bridge.Login(token)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not log in to bridge: %s", err.Error()))
	}
	return nil
}

func updateBridge(bridge *hue.Bridge, ct ColorTemperature, config *PhluxConfig) error {
	err := authenticate(bridge, config)
	if err != nil {
		return err
	}
	log.Println("Logged in to bridge")
	lights, err := bridge.GetAllLights()
	if err != nil {
		return errors.New(fmt.Sprintf("Error getting lights: %s", err.Error()))
	}
	log.Printf("Found %d lights\n", len(lights))
	for _, light := range lights {
		updateLight(light, ct, config)
	}
	return nil
}

func updateLight(light hue.Light, ct ColorTemperature, config *PhluxConfig) {
	log.Printf("Light %d: %s (%s)\n", light.Index, light.Name, light.Type)
	if supportsColorTemp(light) {
		log.Printf("  CT range: %d-%d\n", light.Capabilities.Control.CT.Min, light.Capabilities.Control.CT.Max)
		newCt := ct.TranslateForLight(light)
		log.Printf("  Setting CT to %d\n", newCt)
		light.SetState(hue.LightState{
			On:             light.State.On,
			CT:             newCt,
			TransitionTime: config.TransitionTime,
		})
	}
}
