package v1

import (
	"github.com/gin-gonic/gin"
	"memorial_app_server/service/state"
)

func CurrentState(c *gin.Context) {
	resp := make(map[string]state.State)
	// get all chains
	cluster := state.Chains
	for uid, chain := range *cluster {
		resp[uid] = *chain.GetLastState()
	}
	c.JSON(200, resp)
}

func CurrentChains(c *gin.Context) {
	c.JSON(200, state.Chains)
}

func UseTestRouter(g *gin.RouterGroup) {
	sg := g.Group("/test")
	sg.GET("/state", CurrentState)
	sg.GET("/chains", CurrentChains)
}
