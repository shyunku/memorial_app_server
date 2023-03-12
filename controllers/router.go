package controllers

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"memorial_app_server/configs"
	"memorial_app_server/controllers/v1"
	"memorial_app_server/log"
)

func ping(c *gin.Context) {
	c.String(200, "pong")
}

func SetupRouter() *gin.Engine {
	gin.DefaultWriter = &log.GlobalLogger
	gin.DefaultErrorWriter = &log.GlobalLogger

	// initialize oauth
	v1.InitializeGoogleOauth()

	// setting cors
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}

	r := gin.Default()
	r.Use(cors.New(config))
	r.GET("/ping", ping)

	v1.UseRouterV1(r)
	return r
}

func RunGin() {
	log.Infof("Starting server on port on %d...", configs.AppServerPort)
	r := SetupRouter()
	if err := r.Run(fmt.Sprintf(":%d", configs.AppServerPort)); err != nil {
		log.Error(err)
	}
}
