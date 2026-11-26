package network

import (
	"fmt"
	"math/rand/v2"
	"net"
)

func (n *Network) IsPortAvailable(port int) (bool, error) {
	if port < minPort || port > maxPort {
		return false, fmt.Errorf("port must be between %d and %d (inclusive)", minPort, maxPort)
	}

	// port used by other processes
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return false, fmt.Errorf("port `%d` is already in use", port)
	}
	_ = ln.Close()
	return true, nil
}

func (n *Network) GetAvailablePort() int {
	for i := 0; i < 100; i++ {
		port := rand.IntN(maxPort-minPort+1) + minPort
		if _, err := n.IsPortAvailable(port); err == nil {
			return port
		}
	}
	return -1 // If you actually get this, Good luck! You'll need it
}
