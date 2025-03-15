package utils

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type RateLimit struct {
	Capacity            int `yaml:"capacity"`
	EmptyRatePerSecond  int `yaml:"empty_rate_per_second"`
	RefillRatePerSecond int `yaml:"refill_rate_per_second"`
}

// indivisual endpoint tracking
type resource struct {
	Name           string               `yaml:"name"`
	Endpoint       string               `yaml:"endpoint"`
	DestinationURL string               `yaml:"destination_url"`
	RateLimits     map[string]RateLimit `yaml:"rate_limits"`
}

type configuration struct {

	// server info
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}

	// list of all resources
	Resources []resource
}

// constructor to get configuration from data
func NewConfiguration(filePath string) *configuration {
	var cfg configuration

	// read yaml
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("unable to read config %v", err)
	}

	// decoding the yaml data
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("unable to load config %v", err)
	}

	return &cfg
}
