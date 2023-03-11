package database

type UserEntity struct {
	UserId          *string `db:"uid" json:"userId"`
	AuthId          *string `db:"auth_id" json:"authId"`
	AuthEncryptedPw *string `db:"auth_encrypted_pw" json:"-"`
	GoogleAuthId    *string `db:"google_auth_id" json:"googleAuthId"`
	GoogleEmail     *string `db:"google_email" json:"googleEmail"`
}
