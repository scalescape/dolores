package cld

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/server/cerr"
)

const (
	StagingEnv    Environment = "staging"
	ProductionEnv Environment = "production"
	GlobalEnv     Environment = "global"
	DemoEnv       Environment = "demo"
	EnvID         contextKey  = "env"
)

type (
	Environment string
	contextKey  string
)

func (e Environment) String() string { return string(e) }

func (e Environment) Valid() error {
	if e != StagingEnv && e != ProductionEnv && e != DemoEnv {
		return cerr.ErrInvalidEnvironment
	}
	return nil
}

func ParsePathEnv(r *http.Request) (Environment, error) {
	environment, ok := mux.Vars(r)[string(EnvID)]
	if !ok || environment == "" {
		return "", cerr.ErrInvalidEnvironment
	}
	switch environment {
	case "staging":
		return StagingEnv, nil
	case "demo":
		return DemoEnv, nil
	case "production":
		return ProductionEnv, nil
	default:
		log.Error().Msgf("failed to parse env from path: %s", environment)
		return "", cerr.ErrInvalidEnvironment
	}
}

func ParseQueryEnv(r *http.Request) (Environment, error) {
	env := r.URL.Query().Get(string(EnvID))
	switch env {
	case "staging":
		return StagingEnv, nil
	case "production":
		return ProductionEnv, nil
	case "demo":
		return DemoEnv, nil
	default:
		log.Error().Msgf("failed to parse env from query: %s", env)
		return "", cerr.ErrInvalidEnvironment
	}
}
