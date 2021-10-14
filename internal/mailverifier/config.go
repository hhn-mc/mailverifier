package mailverifier

import (
	_ "embed"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed config.default.yaml
var defaultConfig []byte

func init() {
}

type Config struct {
	EmailRegex             string         `yaml:"email_regex"`
	VerificationCodeLength int            `yaml:"verification_code_length"`
	EmailValidityDuration  string         `yaml:"email_validity_duration"`
	MaxEmailTries          int            `yaml:"max_email_tries"`
	API                    APIConfig      `yaml:"api"`
	Email                  EmailConfig    `yaml:"email"`
	Database               DatabaseConfig `yaml:"database"`
}

type APIConfig struct {
	Bind string `yaml:"bind"`
}

type EmailConfig struct {
	Host     string `yaml:"host"`
	SMTPHost string `yaml:"smtp_host"`
	Email    string `yaml:"email"`
	Alias    string `yaml:"alias"`
	Identity string `yaml:"identity"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func CreateConfigIfNotExist(path string) error {
	if _, err := os.Stat(path); err == nil {
		return err
	}

	return ioutil.WriteFile(path, defaultConfig, 0644)
}

func LoadConfig(path string) (Config, error) {
	bb, err := ioutil.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
