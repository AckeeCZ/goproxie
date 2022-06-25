package util

import (
	"fmt"
	"net"
)

func IsPortFree(port int) bool {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err == nil {
		listener.Close()
		return true
	}
	return false
}
