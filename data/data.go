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

// ErrUserInactive ...
var ErrUserInactive = errors.New("user is inactive")

// ErrForcePasswordChange ...
var ErrForcePasswordChange = errors.New("force password change")

// ErrRoleNotFound ...
var ErrRoleNotFound = errors.New("role not found")

// ErrCollectionAlreadyExists ...
var ErrCollectionAlreadyExists = errors.New("collection already exists")

// ErrCollectionNotFound ...
var ErrCollectionNotFound = errors.New("collection not found")

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

// GetUsers ...
func (m *MongoDB) GetUsers() ([]model.User, error) {
	sess := m.New()
	defer sess.Close()

	var u []model.User

	err := sess.DB("florence").C("users").Find(bson.M{}).All(&u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// GetUser ...
func (m *MongoDB) GetUser(email string) (model.User, error) {
	sess := m.New()
	defer sess.Close()

	var u model.User

	err := sess.DB("florence").C("users").Find(bson.M{"_id": email}).One(&u)
	if err != nil {
		if err == mgo.ErrNotFound {
			return model.User{}, ErrUserNotFound
		}
		return model.User{}, err
	}

	return u, nil
}

// GetRole ...
func (m *MongoDB) GetRole(role string) (model.Role, error) {
	sess := m.New()
	defer sess.Close()

	var r model.Role

	err := sess.DB("florence").C("roles").Find(bson.M{"_id": role}).One(&r)
	if err != nil {
		if err == mgo.ErrNotFound {
			return model.Role{}, ErrRoleNotFound
		}
		return model.Role{}, err
	}

	return r, nil
}

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

// ChangePassword ...
func (m *MongoDB) ChangePassword(email, old, new string) error {
	u, err := m.GetUser(email)
	if err != nil {
		return err
	}

	sess := m.New()
	defer sess.Close()

	if !u.Active {
		return ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(old))
	if err != nil {
		return ErrInvalidPassword
	}

	b, err := bcrypt.GenerateFromPassword([]byte(new), 0)
	if err != nil {
		return err
	}

	err = sess.DB("florence").C("users").Update(bson.M{"_id": email}, bson.M{"$set": bson.M{"password": b, "force_password_change": false}})
	if err != nil {
		return err
	}

	return nil
}

// ValidateLogin ...
func (m *MongoDB) ValidateLogin(email, password string) (string, error) {
	u, err := m.GetUser(email)
	if err != nil {
		return "", err
	}

	sess := m.New()
	defer sess.Close()

	if !u.Active {
		return "", ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidPassword
	}

	if u.ForcePasswordChange {
		return "", ErrForcePasswordChange
	}

	token, err := GenerateRandomString(32)
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
func (m *MongoDB) LoadUser(token string) (model.User, error) {
	sess := m.New()
	defer sess.Close()

	var t model.Token
	err := sess.DB("florence").C("tokens").Find(bson.M{"_id": token}).One(&t)
	if err != nil {
		return model.User{}, ErrInvalidToken
	}

	return m.GetUser(t.Email)
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
