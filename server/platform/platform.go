package platform

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Service struct {
	Store
	log                zerolog.Logger
	DefaultManagedZone string
}

func (s Service) FetchProject(ctx context.Context, oid string, env string) (Project, error) {
	return s.fetchProject(ctx, oid, env)
}

func NewService(defZone string, st Store) Service {
	logger := log.Output(zerolog.NewConsoleWriter()).With().
		Str("service", "platform").
		Logger()
	return Service{DefaultManagedZone: defZone, Store: st, log: logger}
}
