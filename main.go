package main

import (
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
	bindAddr := ":8080"
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

	router.Methods("POST").Path("/login").HandlerFunc(floServer.Login)

	router.Methods("GET").Path("/master/{uri:.*}").Handler(authMw(floServer.MasterData))
	router.Methods("POST").Path("/collections").Handler(authMw(floServer.CreateCollection))

	log.Debug("starting http server", log.Data{"bind_addr": bindAddr})
	if err := srv.ListenAndServe(); err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}
}

func initTest(db *data.MongoDB) {
	sess := db.New()
	defer sess.Close()

	b, err := bcrypt.GenerateFromPassword([]byte("baz"), 0)
	if err != nil {
		panic(err)
	}

	u := model.User{Email: "foo@bar.com", Password: b, Created: time.Now()}
	_, err = sess.DB("florence").C("users").Upsert(bson.M{"_id": "foo@bar.com"}, u)
	if err != nil {
		panic(err)
	}
}
