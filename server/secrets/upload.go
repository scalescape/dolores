package secrets

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/scalescape/dolores/server/cerr"
	"github.com/scalescape/dolores/server/lib"
)

type uploadRequest struct {
	Environment string `json:"environment"`
	Data        string `json:"data"`
	Name        string `json:"name"`
	decodedData []byte `json:"-"`
}

func (r *uploadRequest) Valid() error {
	if r.Environment != "staging" && r.Environment != "production" {
		return cerr.ErrInvalidEnvironment
	}
	var err error
	r.decodedData, err = base64.StdEncoding.DecodeString(r.Data)
	if err != nil {
		return err
	}
	if r.Name == "" {
		return cerr.ErrInvalidSecretRequest
	}
	return nil
}

type uploadResponse struct{}

func Upload(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &uploadRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "invalid request")
			return
		}
		if err := req.Valid(); err != nil {
			lib.WriteError(w, http.StatusBadRequest, err, "invalid request")
			return
		}
		err := svc.UploadSecret(r.Context(), *req)
		if err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to upload secret")
			return
		}
		var resp uploadResponse
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			lib.WriteError(w, http.StatusInternalServerError, err, "failed to encode result")
			return
		}
	}
}
