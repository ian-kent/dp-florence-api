package data

import (
	"errors"
	"time"

	"github.com/ONSdigital/dp-florence-api/data/model"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ErrCollectionAlreadyExists ...
var ErrCollectionAlreadyExists = errors.New("collection already exists")

// ErrCollectionNotFound ...
var ErrCollectionNotFound = errors.New("collection not found")

// GetCollection ...
func (m *MongoDB) GetCollection(id string) (model.Collection, error) {
	sess := m.New()
	defer sess.Close()

	var r model.Collection

	err := sess.DB("florence").C("collections").Find(bson.M{"_id": id}).One(&r)
	if err != nil {
		if err == mgo.ErrNotFound {
			return model.Collection{}, ErrCollectionNotFound
		}
		return model.Collection{}, err
	}

	return r, nil
}

// ListCollections ...
func (m *MongoDB) ListCollections() ([]model.Collection, error) {
	sess := m.New()
	defer sess.Close()

	var r []model.Collection

	err := sess.DB("florence").C("collections").Find(bson.M{}).All(&r)
	if err != nil {
		return []model.Collection{}, err
	}

	return r, nil
}

// CreateCollectionEvent ...
func (m *MongoDB) CreateCollectionEvent(event, collectionID, email string) error {
	sess := m.New()
	defer sess.Close()

	c := model.CollectionEvent{
		Email:        email,
		CollectionID: collectionID,
		Created:      time.Now(),
	}

	err := sess.DB("florence").C("collection_events").Insert(&c)
	return err
}

// CreateCollection ...
func (m *MongoDB) CreateCollection(name, publishType string, publishDate *time.Time, owner, releaseURI string, teams []string) (string, error) {
	sess := m.New()
	defer sess.Close()

	n, err := sess.DB("florence").C("collections").Find(bson.M{"name": name, "published": false}).Count()
	if err != nil {
		return "", err
	}

	if n > 0 {
		return "", ErrCollectionAlreadyExists
	}

	id, err := GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	c := model.Collection{
		ID:              id,
		Name:            name,
		PendingDeletes:  []interface{}{},
		ReleaseURI:      releaseURI,
		Type:            publishType,
		PublishDate:     publishDate,
		CollectionOwner: owner,
		Teams:           []interface{}{},
		Published:       false,
	}

	err = sess.DB("florence").C("collections").Insert(&c)
	if err != nil {
		return "", err
	}

	return id, nil
}
