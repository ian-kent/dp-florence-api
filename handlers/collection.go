package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ONSdigital/dp-florence-api/auth"
	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/go-ns/log"
	"github.com/gorilla/mux"
)

type createCollectionInput struct {
	CollectionOwner string        `json:"collectionOwner"`
	Name            string        `json:"name"`
	PendingDeletes  []interface{} `json:"pendingDeletes"`
	PublishDate     *time.Time    `json:"publishDate"`
	ReleaseURI      string        `json:"releaseUri"`
	Teams           []interface{} `json:"teams"`
	Type            string        `json:"type"`
}

type createCollectionOutput struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Type                  string                        `json:"type"`
	Teams                 []interface{}                 `json:"teams"`
	ApprovalStatus        string                        `json:"approvalStatus"`
	PublishComplete       bool                          `json:"publishComplete"`
	IsEncrypted           bool                          `json:"isEncrypted"`
	PendingDeletes        []interface{}                 `json:"pendingDeletes"`
	CollectionOwner       string                        `json:"collectionOwner"`
	TimeseriesImportFiles []interface{}                 `json:"timeseriesImportFiles"`
	Events                []createCollectionEventOutput `json:"events"`
}

type createCollectionEventOutput struct {
	Date  time.Time `json:"date"`
	Type  string    `json:"type"`
	Email string    `json:"email"`
}

type getCollectionOutput struct {
	ID                    string                        `json:"id"`
	Name                  string                        `json:"name"`
	Type                  string                        `json:"type"`
	Teams                 []interface{}                 `json:"teams"`
	ApprovalStatus        string                        `json:"approvalStatus"`
	InProgress            []interface{}                 `json:"inProgress"`
	Complete              []interface{}                 `json:"complete"`
	Reviewed              []interface{}                 `json:"reviewed"`
	PendingDeletes        []interface{}                 `json:"pendingDeletes"`
	Events                []createCollectionEventOutput `json:"events"`
	TimeseriesImportFiles []interface{}                 `json:"timeseriesImportFiles"`
	CollectionOwner       string                        `json:"collectionOwner"`
	PublishDate           *time.Time                    `json:"publishDate,omitempty"`
	PublishComplete       bool                          `json:"publishComplete"`
}

/*
200 response to create collection
{"approvalStatus":"NOT_STARTED","publishComplete":false,"isEncrypted":true,"pendingDeletes":[],"collectionOwner":"PUBLISHING_SUPPORT","timeseriesImportFiles":[],"events":[{"date":"2017-04-24T01:49:08.096Z","type":"CREATED","email":"florence@magicroundabout.ons.gov.uk"}],"id":"test-95ad38cc6b4b5b82c0cb65b38d36b342e696c53b2f8630267fe8f20e0151b84b","name":"test","type":"manual","teams":[]}

200 response to /collectionDetails/{id}
{"inProgress":[],"complete":[],"reviewed":[],"timeseriesImportFiles":[],"approvalStatus":"NOT_STARTED","pendingDeletes":[],"events":[{"date":"2017-04-24T01:49:08.096Z","type":"CREATED","email":"florence@magicroundabout.ons.gov.uk"}],"collectionOwner":"PUBLISHING_SUPPORT","id":"test-95ad38cc6b4b5b82c0cb65b38d36b342e696c53b2f8630267fe8f20e0151b84b","name":"test","type":"manual","teams":[]}
*/

// ListCollections ...
func (s *FloServer) ListCollections(w http.ResponseWriter, req *http.Request) {
	cols, err := s.DB.ListCollections()
	if err != nil {
		log.DebugR(req, "error fetching collection", log.Data{"error": err})
		if err == data.ErrCollectionNotFound {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(500)
		return
	}

	o := []getCollectionOutput{}

	for _, c := range cols {
		o = append(o, getCollectionOutput{
			ID:                    c.ID,
			Name:                  c.Name,
			Type:                  c.Type,
			Teams:                 c.Teams,
			ApprovalStatus:        "NOT_STARTED",
			PendingDeletes:        c.PendingDeletes,
			CollectionOwner:       c.CollectionOwner,
			Events:                []createCollectionEventOutput{},
			TimeseriesImportFiles: []interface{}{},
			InProgress:            []interface{}{},
			Complete:              []interface{}{},
			Reviewed:              []interface{}{},
			PublishDate:           c.PublishDate,
			PublishComplete:       c.Published,
		})
	}

	b, err := json.Marshal(&o)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(b)
}

// ListPublishedCollections ...
func (s *FloServer) ListPublishedCollections(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`[]`))
}

// GetCollection ...
func (s *FloServer) GetCollection(w http.ResponseWriter, req *http.Request) {
	id := mux.Vars(req)["collection_id"]
	if len(id) == 0 {
		w.WriteHeader(400)
		return
	}

	c, err := s.DB.GetCollection(id)
	if err != nil {
		log.DebugR(req, "error fetching collection", log.Data{"error": err})
		if err == data.ErrCollectionNotFound {
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(500)
		return
	}

	o := getCollectionOutput{
		ID:                    c.ID,
		Name:                  c.Name,
		Type:                  c.Type,
		Teams:                 c.Teams,
		ApprovalStatus:        "NOT_STARTED",
		PendingDeletes:        c.PendingDeletes,
		CollectionOwner:       c.CollectionOwner,
		Events:                []createCollectionEventOutput{},
		TimeseriesImportFiles: []interface{}{},
		InProgress:            []interface{}{},
		Complete:              []interface{}{},
		Reviewed:              []interface{}{},
		PublishDate:           c.PublishDate,
		PublishComplete:       c.Published,
	}

	b, err := json.Marshal(&o)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(b)
}

// CreateCollection ...
func (s *FloServer) CreateCollection(w http.ResponseWriter, req *http.Request) {
	b, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.DebugR(req, "error reading body", log.Data{"error": err})
		w.WriteHeader(400)
		return
	}

	u, ok := auth.UserFromContext(req.Context())
	if !ok {
		log.DebugR(req, "user not in context", nil)
		w.WriteHeader(401)
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

	// TODO input.Teams
	id, err := s.DB.CreateCollection(input.Name, input.Type, input.PublishDate, input.CollectionOwner, input.ReleaseURI, []string{})
	if err != nil {
		log.DebugR(req, "error creating collection", log.Data{"error": err})
		w.WriteHeader(500)
		return
	}

	err = s.DB.CreateCollectionEvent("CREATED", id, u.Email)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	r := createCollectionOutput{
		ID:                    id,
		Name:                  input.Name,
		Type:                  input.Type,
		Teams:                 input.Teams,
		ApprovalStatus:        "NOT_STARTED",
		PublishComplete:       false,
		IsEncrypted:           false,
		PendingDeletes:        input.PendingDeletes,
		CollectionOwner:       input.CollectionOwner,
		TimeseriesImportFiles: []interface{}{},
		Events:                []createCollectionEventOutput{},
	}

	b, err = json.Marshal(&r)
	if err != nil {
		log.ErrorR(req, err, nil)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(201)
	w.Write(b)
}
