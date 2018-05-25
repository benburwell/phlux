package main

import (
	"io/ioutil"
	"log"

	"github.com/BurntSushi/xdg"
	"gopkg.in/yaml.v2"
)

const XDG_CONFIG_NAME = "phlux"

type PhluxConfig struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
	Bridges   []struct {
		BridgeID string `yaml:"id"`
		Token    string `yaml:"token"`
	} `yaml:"bridges"`
}

func (c *PhluxConfig) Read() {
	var paths xdg.Paths
	configFile, err := paths.ConfigFile(XDG_CONFIG_NAME)
	if err != nil {
		log.Printf("No config file found: %s\n", err.Error())
		return
	}
	yamlFile, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Printf("Error reading config file %s: %s\n", configFile, err.Error())
		return
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Printf("Error unmarshalling yaml: %s\n", err.Error())
		return
	}
}
