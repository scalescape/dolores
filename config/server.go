package config

import (
	"fmt"
	"net/url"
)

type Server struct {
	Port int
	Host string
}

func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Backend struct {
	URL *url.URL
}
