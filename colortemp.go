package main

import (
	"math"
	"time"

	hue "github.com/benburwell/gohue"
)

type ColorTemperature uint

func getDesiredColorTemperature(at time.Time) ColorTemperature {
	return 1800
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

// Determine whether there is a non-zero color temperature range for the given light
func supportsColorTemp(l hue.Light) bool {
	return l.Capabilities.Control.CT.Max > l.Capabilities.Control.CT.Min
}
