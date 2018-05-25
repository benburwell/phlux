package main

import (
	"errors"
	"io/ioutil"
	"log"
	"strings"

	"github.com/BurntSushi/xdg"
	"gopkg.in/yaml.v2"
)

const XDG_CONFIG_NAME = "phlux"

type BridgeConfig struct {
	BridgeID string `yaml:"id"`
	Token    string `yaml:"token"`
}

type PhluxConfig struct {
	Latitude  float64        `yaml:"latitude"`
	Longitude float64        `yaml:"longitude"`
	Interval  int64          `yaml:"interval"`
	Bridges   []BridgeConfig `yaml:"bridges"`
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

func (c *PhluxConfig) GetBridgeToken(id string) (string, error) {
	lc := strings.ToLower(id)
	for _, bridge := range c.Bridges {
		if strings.ToLower(bridge.BridgeID) == lc {
			return bridge.Token, nil
		}
	}
	return "", errors.New("No token found for bridge " + id)
}

func (c *PhluxConfig) SetBridgeToken(id, token string) {
	for _, bridge := range c.Bridges {
		if bridge.BridgeID == id {
			bridge.Token = token
			return
		}
	}
	c.Bridges = append(c.Bridges, BridgeConfig{
		BridgeID: id,
		Token:    token,
	})
}

func (c *PhluxConfig) Save() (err error) {
	out, err := yaml.Marshal(c)
	if err != nil {
		log.Printf("Error marshalling config: %s\n", err.Error())
	}
	var paths xdg.Paths
	configFile, err := paths.ConfigFile(XDG_CONFIG_NAME)
	if err != nil {
		log.Printf("No config file found: %s\n", err.Error())
		return
	}
	err = ioutil.WriteFile(configFile, out, 0600)
	if err != nil {
		log.Printf("Error writing config file %s: %s\n", configFile, err.Error())
		return
	}
	return
}
