package frontend

import (
	"net/http"
	"sort"

	"github.com/julienschmidt/httprouter"
	"github.com/memprofiler/memprofiler/schema"
	"github.com/memprofiler/memprofiler/server/storage"
)

// FIXME: to GRPC response
func (s *server) computeSessionMetrics(w http.ResponseWriter, r *http.Request, params httprouter.Params) {

	// parse session id
	sessionID, err := storage.SessionIDFromString(params.ByName("session"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// ask storage for session data
	sessionDescription := &storage.SessionDescription{
		ServiceDescription: &schema.ServiceDescription{
			Type:     params.ByName("type"),
			Instance: params.ByName("instance"),
		},
		SessionID: sessionID,
	}

	// get session metrics
	result, err := s.computer.GetSessionMetrics(r.Context(), sessionDescription)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// sort results by InUseBytes rate, since it the most relevant indicator for memory leak
	sort.Slice(result.Locations, func(i, j int) bool {
		// descending order
		return result.Locations[i].Rates[0].Values.InUseBytes > result.Locations[j].Rates[0].Values.InUseBytes
	})

	// dump to JSON
	m := newJSONMarshaler()
	w.Header().Set("Content-Type", "application/json")
	if err = m.Marshal(w, result); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
