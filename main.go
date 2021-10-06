package main

import "log"

func main() {
	configPath = "config.dev.yaml"
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed laoding config from %s; %s", configPath, err)
	}
	log.Fatal(startAPI(cfg))
}
