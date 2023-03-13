package v1

import (
	"database/sql"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"memorial_app_server/database"
	"memorial_app_server/log"
	"net/http"
)

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
			_, err = database.DB.Exec("INSERT INTO user_master (uid, auth_id, auth_encrypted_pw) VALUES (?, ?, ?)",
				uid, body.AuthId, body.EncryptedPassword)
			if err != nil {
				log.Error(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.Status(http.StatusCreated)
		} else {
			log.Error(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
	} else {
		// user already exists on DB
		c.AbortWithError(http.StatusConflict, errors.New("user already exists"))
	}
}

func UseAuthRouter(g *gin.RouterGroup) {
	sg := g.Group("/auth")
	sg.POST("signup", Signup)
}
