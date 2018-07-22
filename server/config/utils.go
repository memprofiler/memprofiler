package config

import (
	"fmt"
	"net"
	"strconv"
)

func validateEndpoint(endpoint string) error {
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint '%s' format: %v", endpoint, err)
	}

	portValue, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid endpoint '%s' port number: %v", endpoint, err)
	}

	if portValue < 1 || portValue > 65535 {
		return fmt.Errorf("invalid endpoint '%s' port number: %v", endpoint, err)
	}

	if len(host) == 0 {
		return fmt.Errorf("invalid endpoint '%s' host name: %v", host, err)
	}

	return nil
}
