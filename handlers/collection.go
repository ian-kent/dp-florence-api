package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/ONSdigital/go-ns/log"
)

type createCollectionInput struct {
	Name string `json:"name"`
}

// CreateCollection ...
func (s *FloServer) CreateCollection(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	var input createCollectionInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		log.DebugR(req, "error unmarshaling data", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	log.DebugR(req, "create collection", log.Data{"collection": input})

	err = s.DB.CreateCollection(input.Name)
	if err != nil {
		log.DebugR(req, "error creating collection", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	w.WriteHeader(201)
}
