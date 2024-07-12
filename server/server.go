package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/config"
	"github.com/scalescape/dolores/server/platform"
	"github.com/scalescape/dolores/server/secrets"
)

type Application struct {
	router *mux.Router
}

func server(appCfg config.Application) (*Application, error) {
	m := mux.NewRouter()
	m.Use(mux.CORSMethodMiddleware(m))

	aesKey := appCfg.DB.AesKey
	db, err := NewDB(appCfg.DB)
	if err != nil {
		return nil, fmt.Errorf("error creating db conn: %w", err)
	}
	plStore := platform.NewStore(db, aesKey)
	plService := platform.NewService(appCfg.DefaultManagedZone, plStore)

	secService := secrets.NewService(secrets.NewStore(db, aesKey), plService)
	m.Handle("/secrets/recipients", secrets.ListRecipients(secService)).Methods(http.MethodGet, http.MethodOptions)
	m.Handle("/environment/{env}/secrets", secrets.List(secService)).Methods(http.MethodGet, http.MethodOptions)
	m.Handle("/secrets", secrets.Upload(secService)).Methods(http.MethodPut, http.MethodOptions)
	m.Handle("/secrets", secrets.Fetch(secService)).Methods(http.MethodGet, http.MethodOptions)

	return &Application{router: m}, nil
}

func NewDB(cfg config.Database) (*sqlx.DB, error) {
	var err error
	db, err := sqlx.Open(cfg.Driver, cfg.URL())
	if err != nil {
		return nil, fmt.Errorf("error opening conn to db: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging db: %w", err)
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(cfg.MaxConnLifetime())
	return db, nil
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Fatal().Msgf("error: %s", debug.Stack())
		}
	}()

	appCfg := config.MustLoadServer()
	app, err := server(appCfg)
	if err != nil {
		log.Error().Msgf("[Main] error creating server: %v", err)
		return
	}
	addr := appCfg.Address()
	log.Info().Msgf("[Main] listening on address %s", addr)
	server := &http.Server{
		ReadTimeout: 2 * time.Second,
		Addr:        addr,
		Handler:     handlers.LoggingHandler(os.Stdout, app.router),
	}
	go func(server *http.Server) {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal().Msgf("[Main] error listening for rerquests on port: %s err: %v\n", addr, err)
		}
	}(server)
	<-watchSignal()
	// stop HTTP server
	ctx, canc := context.WithTimeout(context.Background(), 5*time.Second)
	defer canc()
	if err := server.Shutdown(ctx); err != nil {
		log.Error().Msgf("[Main] error shutting down server\n")
	}
}

func watchSignal() chan os.Signal {
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}
