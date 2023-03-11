package v1

import "errors"

type userDto struct {
	UserId       string `json:"uid"`
	AuthId       string `json:"auth_id"`
	GoogleAuthId string `json:"google_auth_id"`
	GoogleEmail  string `json:"google_email"`
}

func (u userDto) validate() error {
	if u.UserId == "" {
		return errors.New("uid is empty")
	}
	if u.AuthId == "" && (u.GoogleAuthId == "" || u.GoogleEmail == "") {
		return errors.New("auth_id or google_auth_id and google_email are empty")
	}
	return nil
}
