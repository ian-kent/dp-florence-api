package model

import "time"

// User ...
type User struct {
	Email    string    `bson:"_id" json:"email"`
	Password []byte    `bson:"password" json:"password"`
	Created  time.Time `bson:"created" json:"created"`
}

// Token ...
type Token struct {
	Token   string    `bson:"_id" json:"token"`
	Email   string    `bson:"email" json:"email"`
	Created time.Time `bson:"created" json:"created"`
}
