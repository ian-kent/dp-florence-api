package data

import (
	"errors"
	"net/smtp"
	"strings"
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

// ErrUserExists ...
var ErrUserExists = errors.New("user already exists")

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

// CreateUser ...
func (m *MongoDB) CreateUser(creatorID, email, name string) (err error) {
	_, err = m.GetUser(email)
	if err != ErrUserNotFound {
		return err
	}

	sess := m.New()
	defer sess.Close()

	verificationCode, err := GenerateRandomString(32)
	if err != nil {
		return err
	}

	u := model.User{
		ID:                  bson.NewObjectId(),
		Active:              true,
		Created:             time.Now(),
		Email:               email,
		ForcePasswordChange: true,
		Name:                name,
		VerificationCode:    verificationCode,
	}

	err = sess.DB("florence").C("users").Insert(&u)
	if err != nil {
		return err
	}

	err = m.createAuditEvent(creatorID, AuditEventContextUser, u.ID.Hex(), AuditEventUserCreated, AuditReasonNone)
	if err != nil {
		return err
	}

	err = smtp.SendMail("localhost:1025", nil, "florence@magicroundabout.ons.gov.uk", []string{email}, []byte(`http://localhost:8081/florence/index.html?email=`+email+`&verify=`+verificationCode))
	if err != nil {
		return err
	}

	err = m.createAuditEvent(creatorID, AuditEventContextUser, u.ID.Hex(), AuditEventVerificationEmailSent, AuditReasonNone)
	if err != nil {
		return err
	}

	return nil
}

// GetUser ...
func (m *MongoDB) GetUser(email string) (model.User, error) {
	sess := m.New()
	defer sess.Close()

	var u model.User

	err := sess.DB("florence").C("users").Find(bson.M{"email": email}).One(&u)
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

// ChangePassword ...
func (m *MongoDB) ChangePassword(email, old, new string) error {
	var verify bool
	if strings.HasPrefix(email, "<verify>:") {
		verify = true
		email = strings.TrimPrefix(email, "<verify>:")
	}

	u, err := m.GetUser(email)
	if err != nil {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, email, AuditEventPasswordChangeFailed, AuditReasonUserNotFound)
		if err2 != nil {
			return err2
		}
		return err
	}

	sess := m.New()
	defer sess.Close()

	if !u.Active {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventPasswordChangeFailed, AuditReasonUserInactive)
		if err2 != nil {
			return err2
		}
		return ErrUserInactive
	}

	if verify {
		if old != u.VerificationCode {
			return ErrInvalidPassword
		}
	} else {
		err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(old))
		if err != nil {
			err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventPasswordChangeFailed, AuditReasonInvalidPassword)
			if err2 != nil {
				return err2
			}
			return ErrInvalidPassword
		}
	}

	b, err := bcrypt.GenerateFromPassword([]byte(new), 0)
	if err != nil {
		return err
	}

	err = sess.DB("florence").C("users").Update(bson.M{"_id": u.ID}, bson.M{
		"$set":   bson.M{"password": b, "force_password_change": false},
		"$unset": bson.M{"verification_code": ""},
	})
	if err != nil {
		return err
	}

	err = m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventPasswordChangeOK, AuditReasonNone)
	if err != nil {
		return err
	}

	return nil
}

// ValidateUserVerificationCode ...
func (m *MongoDB) ValidateUserVerificationCode(code string) (ok bool, err error) {
	sess := m.New()
	defer sess.Close()

	var u model.User
	err = sess.DB("florence").C("users").Find(bson.M{"verification_code": code}).One(&u)
	if err != nil {
		return false, err
	}

	return true, nil
}

// ValidateLogin ...
func (m *MongoDB) ValidateLogin(email, password string) (string, error) {
	u, err := m.GetUser(email)
	if err != nil {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, email, AuditEventUserLoginFailed, AuditReasonUserNotFound)
		if err2 != nil {
			return "", err2
		}
		return "", err
	}

	sess := m.New()
	defer sess.Close()

	if !u.Active {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventUserLoginFailed, AuditReasonUserInactive)
		if err2 != nil {
			return "", err2
		}
		return "", ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventUserLoginFailed, AuditReasonInvalidPassword)
		if err2 != nil {
			return "", err2
		}
		return "", ErrInvalidPassword
	}

	if u.ForcePasswordChange {
		err2 := m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventUserLoginFailed, AuditReasonPasswordChangeRequired)
		if err2 != nil {
			return "", err2
		}
		return "", ErrForcePasswordChange
	}

	token, err := GenerateRandomString(32)
	if err != nil {
		return "", err
	}

	err = m.createAuditEvent(AuditSystemUser, AuditEventContextUser, u.ID.Hex(), AuditEventUserLoginOK, AuditReasonNone)
	if err != nil {
		return "", err
	}

	err = sess.DB("florence").C("tokens").Insert(model.Token{Email: email, Token: token, Created: time.Now(), LastActive: time.Now()})
	if err != nil {
		return "", err
	}

	return token, nil
}

// LoadUserFromToken ...
func (m *MongoDB) LoadUserFromToken(token string) (model.User, model.Token, error) {
	sess := m.New()
	defer sess.Close()

	var t model.Token
	err := sess.DB("florence").C("tokens").Find(bson.M{"_id": token}).One(&t)
	if err != nil {
		return model.User{}, model.Token{}, ErrInvalidToken
	}

	u, err := m.GetUser(t.Email)
	return u, t, err
}

// UpdateTokenLastActive ...
func (m *MongoDB) UpdateTokenLastActive(token string) error {
	sess := m.New()
	defer sess.Close()

	return sess.DB("florence").C("tokens").Update(bson.M{"_id": token}, bson.M{"$set": bson.M{"last_active": time.Now()}})
}

// SetUserRoles ...
func (m *MongoDB) SetUserRoles(creatorID, email string, roles ...string) error {
	u, err := m.GetUser(email)
	if err != nil {
		return err
	}

	sess := m.New()
	defer sess.Close()

	err = sess.DB("florence").C("users").Update(bson.M{"email": email}, bson.M{"$set": bson.M{"roles": roles}})
	if err != nil {
		return err
	}

	// FIXME store user roles?
	return m.createAuditEvent(creatorID, AuditEventContextUser, u.ID.Hex(), AuditEventUserRolesUpdated, AuditReasonNone)
}
