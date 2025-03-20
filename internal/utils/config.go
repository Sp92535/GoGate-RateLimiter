package utils

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type RateLimit struct {
	Capacity     int    `yaml:"capacity"`
	Rate         string `yaml:"rate"`
	Strategy     string `yaml:"strategy"`
	NoOfRequests int
	TimeDuration time.Duration
}

// indivisual endpoint tracking
type resource struct {
	Name           string                `yaml:"name"`
	Endpoint       string                `yaml:"endpoint"`
	DestinationURL string                `yaml:"destination_url"`
	RateLimits     map[string]*RateLimit `yaml:"rate_limits"`
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

	for _, resource := range cfg.Resources {
		for key, val := range resource.RateLimits {

			reqStr := strings.Split(val.Rate, "/")

			// setting time duration
			timeUnit := reqStr[1][len(reqStr[1])-1]
			temp := strings.TrimSuffix(reqStr[1], string(timeUnit))
			timeValue := 1
			if temp != "" {
				timeValue, err = strconv.Atoi(temp)
				if err != nil {
					log.Fatalf("Error parsing %v", err)
				}
			}

			switch timeUnit {
			case 'h':
				resource.RateLimits[key].TimeDuration = time.Duration(timeValue) * time.Hour
			case 'm':
				resource.RateLimits[key].TimeDuration = time.Duration(timeValue) * time.Minute
			case 's':
				resource.RateLimits[key].TimeDuration = time.Duration(timeValue) * time.Second
			default:
				log.Fatalf("invalid time unit: %c", timeUnit)
			}

			// setting reqs
			reqUnit := reqStr[0][len(reqStr[0])-1]
			if _, err := strconv.Atoi(string(reqUnit)); err == nil {
				reqValue, err := strconv.Atoi(reqStr[0])
				if err != nil {
					log.Fatalf("Error parsing %v", err)
				}
				resource.RateLimits[key].NoOfRequests = reqValue
				continue
			}
			reqValue, err := strconv.Atoi(strings.TrimSuffix(reqStr[0], string(reqUnit)))

			if err != nil {
				log.Fatalf("Error parsing %v", err)
			}

			switch reqUnit {
			case 'M':
				resource.RateLimits[key].NoOfRequests = reqValue * 1000000
			case 'K':
				resource.RateLimits[key].NoOfRequests = reqValue * 1000
			default:
				log.Fatalf("Error parsing %c not allowed", reqUnit)
			}
		}
	}

	return &cfg
}
