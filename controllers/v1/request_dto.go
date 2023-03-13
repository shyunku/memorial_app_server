package v1

type SignupRequestDto struct {
	Username string `json:"username" binding:"required"`
	AuthId   string `json:"auth_id" binding:"required"`
	// should be double-encrypted from raw password
	EncryptedPassword string `json:"encrypted_password" binding:"required"`
}

type SignupWithGoogleAuthRequestDto struct {
	SignupRequestDto
	GoogleAuthId string `json:"google_auth_id" binding:"required"`
}
