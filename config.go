package main

import (
	_ "embed"
	"fmt"
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
	API      apiConfig      `yaml:"api"`
	Email    emailConfig    `yaml:"email"`
	Database databaseConfig `yaml:"database"`
}

type apiConfig struct {
	Bind       string            `yaml:"bind"`
	EmailRegex string            `yaml:"email_regex"`
	Creds      map[string]string `yaml:"username_password"`
}

type emailConfig struct {
	Host     string `yaml:"host"`
	SMTPHost string `yaml:"smtp_host"`
	Email    string `yaml:"email"`
	Identity string `yaml:"identity"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type databaseConfig struct {
	Address  string `yaml:"address"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (cfg databaseConfig) dsn() string {
	return fmt.Sprintf(
		"host='%s' port='%d' user='%s' password='%s' dbname='%s' sslmode=disable",
		cfg.Address,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Database,
	)
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
