package v1

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"memorial_app_server/database"
	"memorial_app_server/log"
	"memorial_app_server/service"
	"net/http"
	"os"
	"time"
)

// Login handle login without Google auth
func Login(c *gin.Context) {
	var body LoginRequestDto
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if user registered in database
	var userEntity database.UserEntity
	if err := database.DB.QueryRowx("SELECT * FROM user_master WHERE auth_id = ?", body.AuthId).StructScan(&userEntity); err != nil {
		if err == sql.ErrNoRows {
			// user not found
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if userEntity.UserId == nil {
		log.Error(errors.New("user_id is nil"))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userId := *userEntity.UserId

	// set auth token with jwt
	authToken, err := createAuthToken(userId)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := saveAuthToken(userId, authToken); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userDto := UserDtoFromEntity(userEntity)
	authDto := NewAuthTokenDto(authToken.AccessToken, authToken.RefreshToken, false)
	authResult := &authResultDto{
		User: userDto,
		Auth: authDto,
	}

	c.JSON(http.StatusOK, authResult)
}

// Signup handle signup without Google auth
func Signup(c *gin.Context) {
	var body SignupRequestDto
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// check if user already registered in database
	var userEntity database.UserEntity
	result := database.DB.QueryRowx("SELECT * FROM user_master WHERE auth_id = ?", body.AuthId)
	err := result.StructScan(&userEntity)
	if err != nil {
		if err == sql.ErrNoRows {
			// create user
			uid := uuid.New().String()
			_, err := database.DB.Exec("INSERT INTO user_master (uid, auth_id, auth_encrypted_pw) VALUES (?, ?, ?)",
				uid, body.AuthId, body.EncryptedPassword)
			if err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			// successfully created user (most standard case)
			result = database.DB.QueryRowx("SELECT * FROM user_master WHERE uid = ?", uid)
			err = result.StructScan(&userEntity)
			if err != nil {
				// error occurred while getting created user data
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			createdUser := UserDtoFromEntity(userEntity)
			c.JSON(http.StatusCreated, createdUser)
		} else {
			// just db error
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		c.AbortWithError(http.StatusConflict, errors.New("user already registered"))
		return
	}
}

func createAuthToken(uid string) (*authTokenDto, error) {
	var err error
	atd := &authTokenDto{}

	// load jwt secret from env
	jwtAccessSecretKey := os.Getenv("JWT_ACCESS_SECRET")
	jwtRefreshSecretKey := os.Getenv("JWT_REFRESH_SECRET")

	// validate jwt secret
	if jwtAccessSecretKey == "" {
		return nil, errors.New("jwt access secret key is empty")
	}

	if jwtRefreshSecretKey == "" {
		return nil, errors.New("jwt refresh secret key is empty")
	}

	// set access token
	atd.AccessToken.ExpiresAt = time.Now().Add(time.Hour * 3).Unix() // 3 hours expiration
	atd.AccessToken.Uuid = uuid.New().String()
	accessTokenClaims := jwt.MapClaims{}
	accessTokenClaims["uid"] = uid
	accessTokenClaims["exp"] = atd.AccessToken.ExpiresAt
	accessTokenClaims["uuid"] = atd.AccessToken.Uuid
	accessTokenClaims["authorized"] = true
	signedAccessClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, accessTokenClaims)
	atd.AccessToken.Token, err = signedAccessClaims.SignedString([]byte(jwtAccessSecretKey))
	if err != nil {
		return nil, err
	}

	// set refresh token
	atd.RefreshToken.ExpiresAt = time.Now().Add(time.Hour * 24 * 7).Unix() // 7 days expiration
	atd.RefreshToken.Uuid = uuid.New().String()
	refreshTokenClaims := jwt.MapClaims{}
	refreshTokenClaims["uid"] = uid
	refreshTokenClaims["exp"] = atd.RefreshToken.ExpiresAt
	refreshTokenClaims["uuid"] = atd.RefreshToken.Uuid
	signedRefreshClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshTokenClaims)
	atd.RefreshToken.Token, err = signedRefreshClaims.SignedString([]byte(jwtRefreshSecretKey))
	if err != nil {
		return nil, err
	}

	atd.isGoogleAuth = false
	return atd, nil
}

func saveAuthToken(uid string, atd *authTokenDto) error {
	accessTokenExpiresUnix := time.Unix(atd.AccessToken.ExpiresAt, 0)
	refreshTokenExpiresUnix := time.Unix(atd.RefreshToken.ExpiresAt, 0)
	now := time.Now()

	if err := service.InMemoryDB.SetExp(atd.AccessToken.Uuid, uid, accessTokenExpiresUnix.Sub(now)); err != nil {
		return err
	}
	if err := service.InMemoryDB.SetExp(atd.RefreshToken.Uuid, uid, refreshTokenExpiresUnix.Sub(now)); err != nil {
		return err
	}
	return nil
}

func UseAuthRouter(g *gin.RouterGroup) {
	sg := g.Group("/auth")
	sg.POST("login", Login)
	sg.POST("signup", Signup)
}
