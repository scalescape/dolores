package lib

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
)

type errMsg struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, code int, err error, msg string, args ...any) {
	w.WriteHeader(code)
	msg = fmt.Sprintf(msg, args...)
	log.Error().Msgf("%s: %v", msg, err)
	resp := errMsg{Message: msg, Error: err.Error()}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error().Msgf("error writing response: %v", err)
	}
}
