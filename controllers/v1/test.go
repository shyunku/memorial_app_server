package v1

import (
	"github.com/gin-gonic/gin"
	"memorial_app_server/service/state"
	"strconv"
)

func StateHash(c *gin.Context) {
// get block number from query
	blockNumber := c.Query("block_number")

	resp := make(map[string]string)
	if blockNumber == "" {
		// get all chains
		cluster := state.Chains
		for uid, chain := range *cluster {
			resp[uid] = chain.GetLastState().Hash().Hex()
		}
	} else {
		bn, _ := strconv.Atoi(blockNumber)
		cluster := state.Chains
		for uid, chain := range *cluster {
			block, _ := chain.GetBlockByNumber(int64(bn))
			resp[uid] = block.State.Hash().Hex()
		}
	}
	c.JSON(200, resp)
}

func State(c *gin.Context) {
	// get block number from query
	blockNumber := c.Query("block_number")

	resp := make(map[string]state.State)
	if blockNumber == "" {
		// get all chains
		cluster := state.Chains
		for uid, chain := range *cluster {
			resp[uid] = *chain.GetLastState()
		}
	} else {
		bn, _ := strconv.Atoi(blockNumber)
		cluster := state.Chains
		for uid, chain := range *cluster {
			block, _ := chain.GetBlockByNumber(int64(bn))
			blockState := block.State
			resp[uid] = *blockState
		}
	}
	c.JSON(200, resp)
}

func Transaction(c *gin.Context) {
	// get block number from query
	blockNumber := c.Query("block_number")

	resp := make(map[string]state.Transaction)
	if blockNumber == "" {
		// get all chains
		cluster := state.Chains
		for uid, chain := range *cluster {
			lastBn := chain.GetLastBlockNumber()
			block, _ := chain.GetBlockByNumber(lastBn)
			tx := block.Tx
			resp[uid] = *tx
		}
	} else {
		bn, _ := strconv.Atoi(blockNumber)
		cluster := state.Chains
		for uid, chain := range *cluster {
			block, _ := chain.GetBlockByNumber(int64(bn))
			tx := block.Tx
			resp[uid] = *tx
		}
	}
	c.JSON(200, resp)
}

func CurrentChains(c *gin.Context) {
	c.JSON(200, state.Chains)
}

func UseTestRouter(g *gin.RouterGroup) {
	sg := g.Group("/test")
	sg.GET("/state", State)
	sg.GET("/chains", CurrentChains)
	sg.GET("/transaction", Transaction)
}
