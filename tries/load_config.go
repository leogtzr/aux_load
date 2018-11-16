package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config ...
type Config struct {
	Database struct {
		Host     string `json:"host"`
		Password string `json:"password"`
	} `json:"database"`
	Host string `json:"host"`
	Port string `json:"port"`
}

func (cfg *Config) String() string {
	return fmt.Sprintf("db=%q, host=%q | port=%q", cfg.Database, cfg.Host, cfg.Port)
}

// LoadConfiguration ...
func LoadConfiguration(file string) (Config, error) {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return Config{}, err
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	if err != nil {
		return Config{}, err
	}
	return config, nil
}

func main() {
	if config, err := LoadConfiguration("config.json"); err == nil {
		fmt.Println(config.String())
	}
}
