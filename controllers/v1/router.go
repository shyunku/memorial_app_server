package v1

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"io"
	"memorial_app_server/service/database"
	"net/http"
	"os"
	"strings"
)

func UseRouterV1(r *gin.Engine) {
	g := r.Group("/v1")
	g.Use(DefaultMiddleware)
	UseAuthRouter(g)
	UseGoogleAuthRouter(g)
	UseTokenRouter(g)
	UseSocketRouter(g)
}

func DefaultMiddleware(c *gin.Context) {
	//log.Debug(c.Request.Method, c.Request.URL.String())
	c.Next()
}

func AuthMiddleware(c *gin.Context) {
	rawToken, err := extractAuthToken(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	jwtAccessSecretKey := os.Getenv("JWT_ACCESS_SECRET")

	var unauthorizedErr error

	// check if token is valid
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtAccessSecretKey), nil
	})
	if err != nil {
		unauthorizedErr = err
	}

	// check if token is expired
	if unauthorizedErr == nil {
		if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
			// token expired
			unauthorizedErr = errors.New("token expired or invalid")
		}
	}

	if unauthorizedErr != nil {
		// try with Google form access token
		url := "https://www.googleapis.com/oauth2/v1/tokeninfo?access_token=" + rawToken
		req, err := http.NewRequest("POST", url, nil)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": unauthorizedErr.Error()})
			return
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": unauthorizedErr.Error()})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "google auth token expired or invalid"})
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// unmarshal body
		var googleTokenInfo GoogleTokenInfo
		err = json.Unmarshal(body, &googleTokenInfo)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// find user by google auth id
		var userEntity database.UserEntity
		err = database.DB.QueryRowx("SELECT * FROM user_master WHERE google_auth_id = ?", googleTokenInfo.UserId).StructScan(&userEntity)
		if err != nil {
			if err == sql.ErrNoRows{
				c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown user"})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}

		c.Set("uid", *userEntity.UserId)
		c.Next()
	} else {
		claims, ok := token.Claims.(jwt.MapClaims)
		if ok && token.Valid {
			userId := claims["uid"].(string)
			c.Set("uid", userId)
			c.Next()
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		}
	}
}

func extractAuthToken(req *http.Request) (string, error) {
	bearer := req.Header.Get("Authorization")
	token := strings.Split(bearer, " ")
	if len(token) != 2 {
		return "", errors.New("invalid token")
	}
	authToken := token[1]
	if len(authToken) == 0 {
		return "", errors.New("token empty")
	}
	return authToken, nil
}
