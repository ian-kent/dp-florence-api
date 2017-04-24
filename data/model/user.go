package model

import "time"

const (
	// PermAdministrator ...
	PermAdministrator = "administrator"

	// PermEditor ...
	PermEditor = "editor"
)

// User ...
type User struct {
	Email               string    `bson:"_id,omitempty" json:"email"`
	Name                string    `bson:"name" json:"name"`
	Password            []byte    `bson:"password" json:"password"`
	Created             time.Time `bson:"created" json:"created"`
	ForcePasswordChange bool      `bson:"force_password_change" json:"force_password_change"`
	Active              bool      `bson:"active" json:"active"`
	Roles               []string  `bson:"roles" json:"roles"`
}

// Token ...
type Token struct {
	Token   string    `bson:"_id,omitempty" json:"token"`
	Email   string    `bson:"email" json:"email"`
	Created time.Time `bson:"created" json:"created"`
}

// Role ...
type Role struct {
	ID   string `bson:"_id" json:"id,omitempty"`
	Name string `bson:"name" json:"name"`

	Permissions map[string]Permission `bson:"permissions" json:"permissions"`
}

// Permission ...
type Permission struct {
}

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
