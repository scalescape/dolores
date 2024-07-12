package secrets

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/scalescape/dolores/server/cerr"
	"github.com/scalescape/dolores/server/cloud/cld"
	"github.com/scalescape/dolores/server/lib"
	"github.com/scalescape/dolores/server/org"
)

type fetchRequest struct {
	Environment cld.Environment `json:"environment"`
	Name        string          `json:"name"`
	orgID       string          `json:"-"`
}

func (r fetchRequest) Valid() error {
	if err := r.Environment.Valid(); err != nil {
		return err
	}
	if r.Name == "" {
		return cerr.ErrInvalidSecretRequest
	}
	return nil
}

type fetchResponse struct {
	Data string `json:"data"`
}

func Fetch(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := fetchRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "invalid request")
			return
		}
		ctx := r.Context()
		var ok bool
		req.orgID, ok = ctx.Value(org.IDKey).(string)
		if !ok || req.orgID == "" {
			lib.WriteError(w, http.StatusBadRequest, cerr.ErrInvalidOrgID, "invalid org ID")
			return
		}
		if err := req.Valid(); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "invalid request")
			return
		}
		data, err := svc.FetchSecret(ctx, req)
		if err != nil && errors.Is(err, cerr.ErrNoSecretFound) {
			lib.WriteError(w, http.StatusNotFound, err, "failed to fetch secrets")
			return
		}
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to fetch secrets")
			return
		}
		resp := fetchResponse{Data: data}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to encode result")
			return
		}
	}
}
