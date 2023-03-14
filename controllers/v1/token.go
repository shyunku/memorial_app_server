package v1

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	"memorial_app_server/log"
	"memorial_app_server/service/database"
	"net/http"
)

func testToken(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "test",
	})
}

func RefreshToken(c *gin.Context) {
	// get refresh token from header
	refreshToken, err := c.Cookie("refreshToken")
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	storedToken, err := database.InMemoryDB.Get(refreshToken)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// re-generate access token
	rawUserId, exists := c.Get("uid")
	if !exists {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userId := rawUserId.(string)

	// check if user exists
	var userEntity database.UserEntity
	if err := database.DB.QueryRowx("SELECT * FROM user_master WHERE uid = ?", userId).StructScan(&userEntity); err != nil {
		if err == sql.ErrNoRows {
			// user not found
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		log.Error(err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	defer (func() {
		// delete original one
		if err := database.InMemoryDB.Del(storedToken); err != nil {
			log.Error(err)
			return
		}
	})()

	authToken, err := createAuthToken(userId)
	if err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := saveRefreshToken(userId, authToken.RefreshToken); err != nil {
		log.Error(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	authDto := NewAuthTokenDto(authToken.AccessToken, authToken.RefreshToken)
	c.JSON(http.StatusOK, authDto)
}

func UseTokenRouter(g *gin.RouterGroup) {
	sg := g.Group("/token")
	sg.Use(AuthMiddleware)
	sg.POST("/refresh", RefreshToken)
}
