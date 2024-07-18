package main

import (
	"context"
	"net/http"

	"github.com/scalescape/dolores/server/org"
)

func authHeadersMiddleware(next http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		orgID := r.Header.Get("Org-Id")
		ctx := context.WithValue(r.Context(), org.IDKey, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
	return http.HandlerFunc(f)
}
