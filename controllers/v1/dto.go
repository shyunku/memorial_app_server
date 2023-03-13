package v1

import "errors"

type userDto struct {
	UserId                string  `json:"uid"`
	Username              *string `json:"username"`
	AuthId                *string `json:"auth_id"`
	ProfileImageUrl       *string `json:"profile_image_url"`
	GoogleAuthId          *string `json:"google_auth_id"`
	GoogleEmail           *string `json:"google_email"`
	GoogleProfileImageUrl *string `json:"google_profile_image_url"`
}

func NewUserDto(userId string, username, authId, profileImageUrl, googleAuthId, googleEmail, googleProfileImageUrl *string) *userDto {
	return &userDto{
		UserId:                userId,
		AuthId:                authId,
		Username:              username,
		ProfileImageUrl:       profileImageUrl,
		GoogleAuthId:          googleAuthId,
		GoogleEmail:           googleEmail,
		GoogleProfileImageUrl: googleProfileImageUrl,
	}
}

func (u userDto) validate() error {
	if u.UserId == "" {
		return errors.New("uid is empty")
	}
	if (u.AuthId == nil || u.Username == nil) && (u.GoogleAuthId == nil || u.GoogleEmail == nil) {
		return errors.New("auth_id and username or google_auth_id and google_email are empty")
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
