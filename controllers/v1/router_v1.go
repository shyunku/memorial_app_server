package v1

import "github.com/gin-gonic/gin"

func UseRouterV1(r *gin.Engine) {
	g := r.Group("/v1")
	UseGoogleAuthRouter(g)
}
