package config

import ()

type config struct {
	Source               string // Source file of the service URLs
	MaxConcurrentThreads int    // Allowed concurrent threads
	HealthCheckFrequency int    // Frequency of health check in seconds
	Port                 string // Port to bind the application
}

// Settings -- exports the configuration
var Settings *config

func init() {

	Settings = &config{
		Source:               "./target.csv",
		MaxConcurrentThreads: 1024,
		HealthCheckFrequency: 600,
		Port:                 ":8080",
	}
}
