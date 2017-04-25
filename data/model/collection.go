package model

import (
	"time"
)

// Collection ...
type Collection struct {
	ID              string        `bson:"_id,omitempty"`
	Name            string        `bson:"name"`
	CollectionOwner string        `bson:"collection_owner"`
	PendingDeletes  []interface{} `bson:"pending_deletes"`
	PublishDate     *time.Time    `bson:"publish_date"`
	ReleaseURI      string        `bson:"release_uri"`
	Teams           []interface{} `bson:"teams"`
	Type            string        `bson:"type"`
	Published       bool          `bson:"published"`
}

// CollectionEvent ...
type CollectionEvent struct {
	ID           string    `bson:"_id,omitempty"`
	Email        string    `bson:"email"`
	CollectionID string    `bson:"collection_id"`
	Created      time.Time `bson:"created"`
}
