package secrets

import (
	"encoding/json"
	"net/http"

	"github.com/scalescape/dolores/server/cerr"
	"github.com/scalescape/dolores/server/lib"
)

type recipientsRequest struct {
	Environment string `json:"environment"`
}

func (r recipientsRequest) valid() error {
	if r.Environment != "staging" && r.Environment != "production" {
		return cerr.ErrInvalidEnvironment
	}
	return nil
}

type recipientsResponse struct {
	Recipients []Recipient `json:"recipients"`
}

func ListRecipients(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req recipientsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "failed to decode request")
			return
		}
		if err := req.valid(); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "invalid request")
			return
		}
		data, err := svc.ListRecipients(r.Context(), req.Environment)
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to list recipients")
			return
		}
		resp := recipientsResponse{Recipients: data}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to encode result")
			return
		}
	}
}
