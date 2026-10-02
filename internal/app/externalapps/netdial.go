package externalapps

import (
	"net"
	"time"
)

func netDialTimeout(address string, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", address, timeout)
}
