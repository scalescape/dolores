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
	"github.com/scalescape/go-metrics"
)

type httpCli interface {
	Do(*http.Request) (*http.Response, error)
}

type Proxy struct {
	server  *http.Server
	backend *url.URL
	cli     httpCli
}

func GenericHandler(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		log.Trace().Msgf("forwarding request: %s headers: %d", r.URL, len(r.Header))
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(f)
}

func (p *Proxy) Start() error {
	addr := p.server.Addr
	log.Info().Msgf("[Proxy] listening on %s monitoring backend (%s) with reverse proxy", addr, p.backend.String())
	go func(server *http.Server) {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal().Msgf("[Proxy] error listening for rerquests on port: %s err: %v\n", addr, err)
		}
	}(p.server)

	<-watchSignal()
	// stop HTTP server
	ctx, canc := context.WithTimeout(context.Background(), 5*time.Second)
	defer canc()
	if err := p.server.Shutdown(ctx); err != nil {
		log.Error().Msgf("[Proxy] error shutting down server\n")
	}
	return nil
}

func watchSignal() chan os.Signal {
	stop := make(chan os.Signal, 2)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	return stop
}

func NewProxy(cfg config.Server, backend config.Backend, obs metrics.Observer) (*Proxy, error) {
	addr := cfg.Address()
	rp := httputil.NewSingleHostReverseProxy(backend.URL)
	handler := obs.Middleware(GenericHandler(rp))
	server := &http.Server{
		ReadTimeout: 2 * time.Second,
		Addr:        addr,
		Handler:     handler,
	}
	return &Proxy{
		server:  server,
		cli:     http.DefaultClient,
		backend: backend.URL,
	}, nil
}
