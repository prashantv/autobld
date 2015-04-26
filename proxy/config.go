package proxy

import (
	"fmt"
	"strconv"
	"strings"
)

// Config is the configuration to forward a single port.
type Config struct {
	// Port is the port that the file watcher listens on.
	Port int `yaml:"port"`
	// ForwardTo is the port that requests will be proxied to.
	ForwardTo int `yaml:"forwardTo"`
	// Type is the type of proxy to use.
	Type Type `yaml:"type"`
	// HTTPPath is the path on the target server to use as the base.
	HTTPPath string `yaml:"httpPath"`
}

// Parse parses a proxy argument passed to through flags and returns a Config.
func Parse(p string) (Config, error) {
	pConfig := Config{}
	parts := strings.Split(p, ":")
	if lp := len(parts); lp != 2 && lp != 3 {
		return pConfig, fmt.Errorf("proxy port is not in format listen:forward, got %v", p)
	}

	var err error
	if len(parts) == 3 {
		pConfig.Type = stringToType(parts[0])
		parts = parts[1:]
	}

	pConfig.Port, err = strconv.Atoi(parts[0])
	if err != nil {
		return pConfig, err
	}

	targetParts := strings.SplitN(parts[1], "/", 2)
	pConfig.ForwardTo, err = strconv.Atoi(targetParts[0])
	if err != nil {
		return pConfig, err
	}

	if len(targetParts) > 1 {
		if pConfig.Type != HTTP {
			return pConfig, fmt.Errorf("only HTTP proxies can have a target path: %v", p)
		}
		pConfig.HTTPPath = targetParts[1]
	}

	return pConfig, nil
}
