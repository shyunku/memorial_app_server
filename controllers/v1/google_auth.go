package v1

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"io"
	"memorial_app_server/database"
	"memorial_app_server/log"
	"net/http"
	"os"
)

var (
	clientId,
	clientSecret,
	redirectUrl string
	config *oauth2.Config
)

type GoogleOauth2UserInfo struct {
	Email         string `json:"email"`
	Id            string `json:"id"`
	Picture       string `json:"picture"`
	VerifiedEmail bool   `json:"verified_email"`
}

func InitializeGoogleOauth() {
	clientId = os.Getenv("GOOGLE_OAUTH2_CLIENT_ID")
	clientSecret = os.Getenv("GOOGLE_OAUTH2_CLIENT_SECRET")
	redirectUrl = os.Getenv("GOOGLE_OAUTH2_REDIRECT_URL")
	config = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		RedirectURL:  redirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: "https://accounts.google.com/o/oauth2/token",
		},
	}

	if clientId == "" || clientSecret == "" || redirectUrl == "" {
		panic("Missing environment variables for Google OAuth2 configuration")
	}
}

func GoogleOauth2Login(c *gin.Context) {
	// create random token to prevent CSRF
	stateToken := createGoogleOauthState()
	// save token to session
	c.SetCookie("oauthstate", stateToken, 0, "/", "", false, true)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate") // Set Cache-Control header
	url := config.AuthCodeURL(stateToken)
	// redirect to Google's consent page to ask for permission
	c.Redirect(http.StatusMovedPermanently, url)
}

func GoogleOauth2Callback(c *gin.Context) {
	stateToken, err := c.Cookie("oauthstate")
	if err != nil {
		log.Error(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if c.Query("state") != stateToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	token, err := config.Exchange(context.Background(), c.Query("code"))
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	log.Debug("Contents:", contents)

	var googleOauthUserInfo GoogleOauth2UserInfo
	err = json.Unmarshal(contents, &googleOauthUserInfo)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// create user if not exist
	var userEntity database.UserEntity
	var user *userDto
	result := database.DB.QueryRowx("SELECT * FROM user_master WHERE google_auth_id = ?", googleOauthUserInfo.Id)
	err = result.StructScan(&userEntity)
	if err != nil {
		if err == sql.ErrNoRows {
			// create user
			uid := uuid.New().String()
			_, err = database.DB.Exec("INSERT INTO user_master (uid, google_auth_id, google_email) VALUES (?, ?, ?)",
				uid, googleOauthUserInfo.Id, googleOauthUserInfo.Email)
			if err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			user = &userDto{
				UserId:       uid,
				AuthId:       googleOauthUserInfo.Email,
				GoogleAuthId: googleOauthUserInfo.Id,
				GoogleEmail:  googleOauthUserInfo.Email,
			}
		} else {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		user = &userDto{
			UserId:       *userEntity.UserId,
			AuthId:       *userEntity.GoogleEmail,
			GoogleAuthId: *userEntity.GoogleAuthId,
			GoogleEmail:  *userEntity.GoogleEmail,
		}
	}

	if user == nil {
		log.Error("user is nil")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err = user.validate(); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	log.Debug("user:", user)

	c.JSON(http.StatusOK, user)
}

func createGoogleOauthState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func UseGoogleAuthRouter(g *gin.RouterGroup) {
	sg := g.Group("/google_auth")
	sg.GET("login", GoogleOauth2Login)
	sg.GET("login_callback", GoogleOauth2Callback)
}
