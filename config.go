package main

import (
	_ "embed"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

var configPath = "config.yaml"

//go:embed config.yaml
var defaultConfig []byte

func init() {
	if _, err := os.Stat(configPath); err == nil {
		return
	}

	if err := ioutil.WriteFile(configPath, defaultConfig, 0644); err != nil {
		log.Fatalf("Failed to create default config at %s; %s", configPath, err)
	}
}

type config struct {
	API struct {
		Bind string `yaml:"bind"`
	} `yaml:"api"`
	Email struct {
		Host     string `yaml:"host"`
		SMTPHost string `yaml:"smtp_host"`
		Email    string `yaml:"email"`
		Identity string `yaml:"identity"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"email"`
	EmailRegex string `yaml:"email_regex"`
}

func loadConfig() (config, error) {
	bb, err := ioutil.ReadFile(configPath)
	if err != nil {
		return config{}, err
	}

	var cfg config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}
