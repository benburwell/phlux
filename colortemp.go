package main

import (
	"time"
)

type ColorTemperature uint

func getDesiredColorTemperature(at time.Time) ColorTemperature {
	return 1800
}
