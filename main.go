package main

import (
	"net/http"
	"os"
	"strconv"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/ONSdigital/dp-florence-api/auth"
	"github.com/ONSdigital/dp-florence-api/data"
	"github.com/ONSdigital/dp-florence-api/data/model"
	"github.com/ONSdigital/dp-florence-api/handlers"
	"github.com/ONSdigital/go-ns/log"
	"github.com/ONSdigital/go-ns/server"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	bindAddr := ":8082"
	mongoURI := "mongodb://localhost:27017"
	initDB := false

	if v := os.Getenv("BIND_ADDR"); len(v) > 0 {
		bindAddr = v
	}

	if v := os.Getenv("MONGO_URI"); len(v) > 0 {
		mongoURI = v
	}

	if v := os.Getenv("INIT_DB"); len(v) > 0 {
		initDB, _ = strconv.ParseBool(v)
	}

	mongoDB, err := data.NewMongoDB(mongoURI)
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	if initDB {
		initTest(mongoDB)
	}

	floServer := &handlers.FloServer{DB: mongoDB}
	authMw := auth.Middleware(mongoDB)

	router := mux.NewRouter()
	srv := server.New(bindAddr, router)

	// FIXME move to /api
	//root := router.PathPrefix("/")
	var root = router

	root.Methods("POST").Path("/login").HandlerFunc(floServer.Login)
	root.Methods("POST").Path("/password").HandlerFunc(floServer.ChangePassword)

	root.Methods("GET").Path("/master/{uri:.*}").Handler(authMw(floServer.MasterData))

	root.Methods("GET").Path("/publishedCollections").Handler(authMw(floServer.ListPublishedCollections))
	root.Methods("GET").Path("/collections").Handler(authMw(floServer.ListCollections))
	root.Methods("POST").Path("/collection").Handler(authMw(floServer.CreateCollection))
	root.Methods("GET").Path("/collectionDetails/{collection_id}").Handler(authMw(floServer.GetCollection))

	root.Methods("GET").Path("/users").Handler(authMw(floServer.ListUsers))
	root.Methods("POST").Path("/users").Handler(authMw(floServer.CreateUser))
	root.Methods("GET").Path("/teams").Handler(authMw(floServer.ListTeams))
	root.Methods("GET").Path("/permission").Handler(authMw(floServer.Permissions))

	root.Methods("POST").Path("/ping").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// TODO output real stuff here
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"hasSession": true, sessionExpiryDate: "2017-05-01T01:39:27.061Z"}`))
	})

	// root.Methods("GET").Path("/").Handler(authMw(func(w http.ResponseWriter, req *http.Request) {}))
	// root.Methods("GET").Path("/data").Handler(authMw(func(w http.ResponseWriter, req *http.Request) {}))
	// root.Methods("GET").Path("/taxonomy").Handler(authMw(func(w http.ResponseWriter, req *http.Request) {}))

	root.Methods("POST").Path("/clickEventLog").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// TODO ?
	})

	root.NotFoundHandler = authMw(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(404)
	})

	log.Debug("starting http server", log.Data{"bind_addr": bindAddr})
	if err := srv.ListenAndServe(); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}

func initTest(db *data.MongoDB) {
	sess := db.New()
	defer sess.Close()

	b, err := bcrypt.GenerateFromPassword([]byte("Doug4l"), 0)
	if err != nil {
		panic(err)
	}

	r := model.Role{
		ID:   "administrator",
		Name: "Administrator",
		Permissions: map[string]model.Permission{
			model.PermAdministrator: model.Permission{},
		},
	}
	_, err = sess.DB("florence").C("roles").Upsert(bson.M{"_id": "administrator"}, r)
	if err != nil {
		panic(err)
	}

	r = model.Role{
		ID:   "editor",
		Name: "Editor",
		Permissions: map[string]model.Permission{
			model.PermEditor: model.Permission{},
		},
	}
	_, err = sess.DB("florence").C("roles").Upsert(bson.M{"_id": "editor"}, r)
	if err != nil {
		panic(err)
	}

	u := model.User{Email: "florence@magicroundabout.ons.gov.uk", Name: "Florence", Password: b, Created: time.Now(), Active: true, ForcePasswordChange: true, Roles: []string{"administrator", "editor"}}
	_, err = sess.DB("florence").C("users").Upsert(bson.M{"email": "florence@magicroundabout.ons.gov.uk"}, u)
	if err != nil {
		panic(err)
	}
}
