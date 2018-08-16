package web

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/vitalyisaev2/memprofiler/schema"
	"github.com/vitalyisaev2/memprofiler/server/storage"
)

func (s *server) computeSessionMetrics(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	description := &schema.ServiceDescription{
		Type:     params.ByName("type"),
		Instance: params.ByName("instance"),
	}

	// parse session id
	sessionID, err := storage.SessionIDFromString(params.ByName("session"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ask storage for session data
	loader, err := s.storage.NewDataLoader(description, sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// compute metrics for a session
	result, err := s.computer.SessionMetrics(r.Context(), loader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// dump to JSON
	encoder := newJSONEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	if err = encoder.Encode(result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
