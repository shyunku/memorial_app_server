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

type authTokenDto struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiredAt    int64  `json:"expired_at"`
}

func NewAuthTokenDto(accessToken, refreshToken string, expiredAt int64) *authTokenDto {
	return &authTokenDto{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiredAt:    expiredAt,
	}
}

type googleAuthResultDto struct {
	User          *userDto      `json:"user"`
	Auth          *authTokenDto `json:"auth"`
	NewlySignedUp bool          `json:"newly_signed_up"`
}
