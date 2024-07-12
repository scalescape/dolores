package config

import (
	"fmt"
	"log"
	"net/url"

	"github.com/kelseyhightower/envconfig"
)

type Application struct {
	Server
	DB Database
}

var app Application

type Server struct {
	DefaultManagedZone string
	Port               int
	Host               string
}

func (app Application) Address() string {
	return fmt.Sprintf("%s:%d", app.Server.Host, app.Server.Port)
}

func (s Server) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

type Backend struct {
	URL *url.URL
}

func MustLoadServer() Application {
	var errs []error
	if err := envconfig.Process("", &app); err != nil {
		errs = append(errs, err)
	}
	if len(errs) != 0 {
		log.Fatalf("Error loading configuration: %v", errs)
	}
	log.Println("config loaded successfully")
	return app
}
