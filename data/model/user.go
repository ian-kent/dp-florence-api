package model

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

const (
	// PermAdministrator ...
	PermAdministrator = "administrator"

	// PermEditor ...
	PermEditor = "editor"
)

// User ...
type User struct {
	ID                  bson.ObjectId `bson:"_id,omitempty"`
	Email               string        `bson:"email"`
	Name                string        `bson:"name"`
	Password            []byte        `bson:"password"`
	Created             time.Time     `bson:"created"`
	ForcePasswordChange bool          `bson:"force_password_change"`
	Active              bool          `bson:"active"`
	Roles               []string      `bson:"roles"`
	VerificationCode    string        `bson:"verification_code"`
}

// Token ...
type Token struct {
	Token   string    `bson:"_id,omitempty"`
	Email   string    `bson:"email"`
	Created time.Time `bson:"created"`
}

// Role ...
type Role struct {
	ID   string `bson:"_id"`
	Name string `bson:"name"`

	Permissions map[string]Permission `bson:"permissions" json:"permissions"`
}

// Permission ...
type Permission struct {
}
