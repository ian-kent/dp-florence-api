package data

import (
	"errors"
	"time"

	"github.com/ONSdigital/dp-florence-api/data/model"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ErrUserNotFound ...
var ErrUserNotFound = errors.New("user not found")

// ErrInvalidPassword ...
var ErrInvalidPassword = errors.New("invalid password")

// ErrInvalidToken ...
var ErrInvalidToken = errors.New("invalid token")

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

// ValidateLogin ...
func (m *MongoDB) ValidateLogin(email, password string) (string, error) {
	sess := m.New()
	defer sess.Close()

	var u model.User

	err := sess.DB("florence").C("users").Find(bson.M{"email": email}).One(&u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return "", ErrUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	token, err := generateRandomString(32)
	if err != nil {
		return "", err
	}

	err = sess.DB("florence").C("tokens").Insert(model.Token{Email: email, Token: token, Created: time.Now()})
	if err != nil {
		return "", err
	}

	return token, nil
}

// LoadUser ...
func (m *MongoDB) LoadUser(token string) (*model.User, error) {
	sess := m.New()
	defer sess.Close()

	var t model.Token
	err := sess.DB("florence").C("tokens").Find(bson.M{"token": token}).One(&t)
	if err != nil {
		return nil, ErrInvalidToken
	}

	var u model.User
	err = sess.DB("florence").C("users").Find(bson.M{"email": t.Email}).One(&u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// CreateCollection ...
func (m *MongoDB) CreateCollection(name string) error {
	return nil
}
