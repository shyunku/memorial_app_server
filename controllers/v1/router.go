package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
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

	// check if token is valid
	token, err := jwt.Parse(rawToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtAccessSecretKey), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// check if token is expired
	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		// token expired
		c.JSON(
			http.StatusUnauthorized,
			gin.H{"error": "token expired or invalid"},
		)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		userId := claims["uid"].(string)
		c.Set("uid", userId)
		c.Next()
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
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
