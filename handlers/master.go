package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

// MasterData ...
func (s *FloServer) MasterData(w http.ResponseWriter, req *http.Request) {
	uri := mux.Vars(req)["uri"]

	log.DebugR(req, "master data uri", log.Data{"uri": uri})

	m := map[string]interface{}{
		"uri": uri,
	}

	b, err := json.Marshal(&m)
	if err != nil {
		log.DebugR(req, "error marshaling json", log.Data{"error": err})
		w.WriteHeader(500)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.DebugR(req, "error writing response", log.Data{"error": err})
	}
}
