package monitor

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/config"
)

type httpCli interface {
	Do(*http.Request) (*http.Response, error)
}

type Proxy struct {
	server  config.Server
	backend *url.URL
	cli     httpCli
}

func (p Proxy) GenericHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msgf("forwarding request: %s path: %d", r.URL, len(r.Header))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func (p Proxy) Start() error {
	addr := p.server.Address()
	log.Info().Msgf("[Main] listening on address %s", addr)
	rp := httputil.NewSingleHostReverseProxy(p.backend)
	server := &http.Server{
		ReadTimeout: 2 * time.Second,
		Addr:        addr,
		Handler:     p.GenericHandler(rp),
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
	return nil
}

func watchSignal() chan os.Signal {
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}

func NewProxy(cfg config.Server, backend config.Backend) Proxy {
	return Proxy{
		server:  cfg,
		backend: backend.URL,
		cli:     http.DefaultClient}
}
