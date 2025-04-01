// Package flags for custom flag implementation.
package flags

import (
	"errors"
	"net"
)

// NetAddress struct for config.
type NetAddress struct {
	Host string // Host i.e. http://localhost.
	Port string // Port value w/o ":".
}

// String converts struct to address string.
func (a *NetAddress) String() string {
	return a.Host + ":" + a.Port
}

// Set parse address string to struct.
func (a *NetAddress) Set(s string) error {
	host, port, err := net.SplitHostPort(s)

	if err != nil {
		return errors.New("need address in a form host:port")
	}

	if port == "" {
		return errors.New("at least port should be defined")
	}

	if host == "" {
		host = "localhost"
	}

	a.Host = host
	a.Port = port

	return nil
}
