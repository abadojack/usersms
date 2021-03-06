package mocks

import (
	"github.com/tomogoma/go-typed-errors"
	"github.com/tomogoma/usersms/pkg/user"
	"time"
)

type User struct {
	errors.ErrToHTTP

	UpdtRecTkn     string
	UpdtRecUsrUpdt user.UserUpdate
	UpdtUsr        *user.User
	UpdtErr        error

	UsrRecTkn        string
	UsrRecID         string
	UsrRecOffstUpdDt time.Time
	UsrUsr           *user.User
	UsrErr           error
}

func (u *User) Update(token string, update user.UserUpdate) (*user.User, error) {
	u.UpdtRecTkn = token
	u.UpdtRecUsrUpdt = update
	return u.UpdtUsr, u.UpdtErr
}

func (u *User) User(token, ID string, offsetUpdateDate time.Time) (*user.User, error) {
	u.UsrRecTkn = token
	u.UsrRecID = ID
	u.UsrRecOffstUpdDt = offsetUpdateDate
	return u.UsrUsr, u.UsrErr
}
