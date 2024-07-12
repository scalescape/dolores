package secrets

import (
	"encoding/json"
	"net/http"

	"github.com/scalescape/dolores/server/cerr"
	"github.com/scalescape/dolores/server/cloud/cld"
	"github.com/scalescape/dolores/server/lib"
	"github.com/scalescape/dolores/server/org"
)

type listResponse struct {
	Secrets []Secret `json:"secrets"`
}

type listRequest struct {
	environment cld.Environment
	orgID       string
}

func (r listRequest) Valid() error {
	return nil
}

func List(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var req listRequest
		var err error
		req.environment, err = cld.ParsePathEnv(r)
		if err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "failed to parse environment")
			return
		}
		var ok bool
		req.orgID, ok = ctx.Value(org.IDKey).(string)
		if !ok {
			lib.WriteError(w, http.StatusBadRequest, cerr.ErrInvalidOrgID, "invalid org ID")
			return
		}
		if err := req.Valid(); err != nil {
			lib.WriteError(w, http.StatusBadRequest, cerr.ErrInvalidOrgID, "invalid org ID")
			return
		}
		data, err := svc.ListSecret(ctx, req)
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to list secrets")
			return
		}
		if len(data) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		resp := listResponse{data}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to encode result")
			return
		}
	}
}
