package main

import (
	"log"
	"math"
	"time"

	hue "github.com/benburwell/gohue"
	"github.com/cpucycle/astrotime"
)

type ColorTemperature uint

const LATITUDE = 42.348333
const LONGITUDE = 71.1675

// Get the correct color temperature for a given time. If the time is between
// sunrise and sunset, a high CT is selected. Otherise, during night, a low
// color temperature is used.
//
// TODO: This is a naive approach and should be revisited
func getDesiredColorTemperature(t time.Time) ColorTemperature {
	log.Printf("Calculating sunrise/sunset for: %s\n", t.Format(time.RFC3339))
	sunrise := astrotime.CalcSunrise(t, LATITUDE, LONGITUDE)
	sunset := astrotime.CalcSunset(t, LATITUDE, LONGITUDE)
	log.Printf("Sunrise: %s, Sunset: %s\n", sunrise.Format(time.RFC3339), sunset.Format(time.RFC3339))
	if t.After(sunrise) && t.Before(sunset) {
		log.Println("Daytime, setting high CT")
		return 6500
	} else {
		log.Println("Nighttime, setting low CT")
		return 1800
	}
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
func (ct ColorTemperature) TranslateForLight(light hue.Light) uint16 {
	divisor := 12.968
	scaled := float64(ct) / divisor
	inverted := 500 - scaled + 153
	min := float64(light.Capabilities.Control.CT.Min)
	max := float64(light.Capabilities.Control.CT.Max)
	return uint16(math.Max(min, math.Min(max, inverted)))
}

// Determine whether there is a non-zero color temperature range for the given light
func supportsColorTemp(l hue.Light) bool {
	return l.Capabilities.Control.CT.Max > l.Capabilities.Control.CT.Min
}
