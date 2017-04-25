package data

import (
	"gopkg.in/mgo.v2"
)

// MongoDB ...
type MongoDB struct {
	*mgo.Session
}

// NewMongoDB ...
func NewMongoDB(url string) (*MongoDB, error) {
	session, err := mgo.Dial(url)
	if err != nil {
		return nil, err
	}

	return &MongoDB{session}, nil
}
