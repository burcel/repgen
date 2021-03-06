package core

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

const configFileName = "config.yaml"

type ConfigBase struct {
	// Server version
	Version string `yaml:"version"`
	// Server config
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	// PostgreSQL database config
	Postgresql struct {
		Host               string `yaml:"host"`
		Port               string `yaml:"port"`
		User               string `yaml:"user"`
		Password           string `yaml:"password"`
		Database           string `yaml:"database"`
		MaxIdleConnections int    `yaml:"max_idle_conns"`
		MaxOpenConnections int    `yaml:"max_open_conns"`
	} `yaml:"postgresql"`
}

var Config *ConfigBase

func InitializeConfig() {
	// Read config file
	content, err := ioutil.ReadFile(configFileName)
	if err != nil {
		panic(err)
	}
	// Parse YAML file
	err = yaml.Unmarshal(content, &Config)
	if err != nil {
		panic(err)
	}
	log.Printf("Backend version: %s", Config.Version)
}
