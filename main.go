package main

import (
	"log"
	"math"
	"os"
	"time"

	hue "github.com/benburwell/gohue"
)

const username = "phlux"

func main() {
	bridges, err := hue.FindBridges()
	if err != nil {
		log.Fatalf("Error finding bridges: %s\n", err.Error())
	}
	log.Printf("Found %d bridge(s)\n", len(bridges))
	for _, bridge := range bridges {
		log.Printf("Bridge: %s\n", bridge.IPAddress)
		//username, err := bridge.CreateUser(username)
		//if err != nil {
		//	panic("Could not create user on bridge")
		//}
		//fmt.Printf("Made user %s for bridge %s\n", username, bridge.IPAddress)
		err = bridge.Login(os.Getenv("HUE_LOGIN"))
		if err != nil {
			log.Fatalf("Could not log in to bridge: %s", err.Error())
		}
		log.Println("Logged in to bridge")
		lights, err := bridge.GetAllLights()
		if err != nil {
			log.Fatalf("Error getting lights: %s\n", err.Error())
		}
		log.Printf("Found %d lights\n", len(lights))
		for _, light := range lights {
			log.Printf("Light %d: %s (%s)\n", light.Index, light.Name, light.Type)
			if supportsColorTemp(light) {
				log.Printf("  CT range: %d-%d\n", light.Capabilities.Control.CT.Min, light.Capabilities.Control.CT.Max)
				if light.Index == 8 {
					newCt := translateCtForLight(getDesiredColorTemperature(time.Now()), light)
					log.Printf("  Setting CT to %d\n", newCt)
					light.SetState(hue.LightState{
						On: light.State.On,
						CT: newCt,
					})
				}
			}
		}
	}
}

// Determine whether there is a non-zero color temperature range for the given light
func supportsColorTemp(l hue.Light) bool {
	return l.Capabilities.Control.CT.Max > l.Capabilities.Control.CT.Min
}

// Translate a desired color temperature in Kelvins to a value comprehensible
// by a Hue luminaire. According to Hue documentation, the value 153
// corresponds to 6500K, and 500 corresponds to 2000K. Using these known
// values, we'll create a mapping between spaces, and additionally limit the
// resulting value by the range that the luminaire supports.
//
//   153 <=> 6500K
//   500 <=> 2000K
//   =============
//   347     4500
func translateCtForLight(ct ColorTemperature, light hue.Light) uint16 {
	divisor := 12.968
	scaled := float64(ct) / divisor
	inverted := 500 - scaled + 153
	min := float64(light.Capabilities.Control.CT.Min)
	max := float64(light.Capabilities.Control.CT.Max)
	return uint16(math.Max(min, math.Min(max, inverted)))
}
