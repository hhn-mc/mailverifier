package main

import (
	_ "embed"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

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
	EmailRegex             string         `yaml:"email_regex"`
	VerificationCodeLength int            `yaml:"verification_code_length"`
	EmailValidityDuration  string         `yaml:"email_validity_duration"`
	MaxEmailTries          int            `yaml:"max_email_tries"`
	API                    apiConfig      `yaml:"api"`
	Email                  emailConfig    `yaml:"email"`
	Database               databaseConfig `yaml:"database"`
}

type apiConfig struct {
	Bind string `yaml:"bind"`
}

type emailConfig struct {
	Host     string `yaml:"host"`
	SMTPHost string `yaml:"smtp_host"`
	Email    string `yaml:"email"`
	Alias    string `yaml:"alias"`
	Identity string `yaml:"identity"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type databaseConfig struct {
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func loadConfig(path string) (config, error) {
	bb, err := ioutil.ReadFile(path)
	if err != nil {
		return config{}, err
	}

	var cfg config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return config{}, err
	}

	return cfg, nil
}
