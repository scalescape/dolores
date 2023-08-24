package monitor

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type Proxy struct {
	cfg Config
}

type Config struct {
	Port int
	Host string
}

func (s Config) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func router() *mux.Router {
	m := mux.NewRouter()
	//m.Use(mux.MiddlewareFunc(contentWriter))
	m.Use(mux.CORSMethodMiddleware(m))
	//m.Use(mux.MiddlewareFunc(accessController))
	m.HandleFunc("/.*", GenericHandler)
	return m
}

func GenericHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/ping"))
	log.Info().Msgf("got request: %s", r.URL)
}

func (p Proxy) Start() {
	addr := p.cfg.Address()
	log.Info().Msgf("[Main] listening on address %s", addr)
	server := &http.Server{
		ReadTimeout: 2 * time.Second,
		Addr:        addr,
		Handler:     http.HandlerFunc(GenericHandler),
	}
	go func(server *http.Server) {
		err := server.ListenAndServe()
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

func NewProxy(cfg Config) Proxy { return Proxy{cfg} }
