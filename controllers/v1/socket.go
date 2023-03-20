package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"math/big"
	"memorial_app_server/log"
	"memorial_app_server/service/state"
)

type socketHandler func(conn *websocket.Conn, uid string, data interface{}) (interface{}, error)

var (
	socketHandlers = map[string]socketHandler{
		"test":        test,
		"transaction": handleTransaction,
	}
	socketBundles = map[string]*UserSocketBundle{}
)

func test(conn *websocket.Conn, uid string, data interface{}) (interface{}, error) {
	// data to string
	str, ok := data.(string)
	if !ok {
		return nil, errors.New("invalid data type")
	}
	return str, nil
}

func handleTransaction(conn *websocket.Conn, uid string, data interface{}) (interface{}, error) {
	request, ok := data.(*TxSocketRequest)
	if !ok {
		log.Error("Invalid data type")
		return nil, fmt.Errorf("invalid request: check format")
	}

	userChain := state.Chains.GetChain(uid)

	// check if targetBlockNumber is valid
	targetWaitingBlockNumber, ok := new(big.Int).SetString(request.TargetBlockNumber, 10)
	if !ok {
		log.Error("Invalid target block number")
		return nil, fmt.Errorf("invalid block number: couldn't be parsed")
	}

	waitingBlockNumber := userChain.GetWaitingBlockNumber()
	if targetWaitingBlockNumber.Cmp(waitingBlockNumber) != 0 {
		// different block number
		log.Error("Invalid target block number")
		return nil, fmt.Errorf("invalid block number: waiting for block #%s", waitingBlockNumber.String())
	}

	// check if transaction is valid
	tx := state.NewTransaction(request.From, request.Type, request.Timestamp, request.Content)
	if tx.From != uid {
		log.Error("Invalid transaction")
		return nil, fmt.Errorf("invalid source: trying to apply transaction as other user")
	}
	if err := tx.Validate(); err != nil {
		log.Error("Invalid transaction")
		return nil, fmt.Errorf("invalid request: %s", err.Error())
	}

	// apply transaction
	if err := userChain.ApplyTransaction(tx); err != nil {
		log.Error("Error during applying transaction")
		return nil, fmt.Errorf("failed to apply transaction: %s", err.Error())
	}

	return nil, nil
}

func SocketV1(c *gin.Context) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error("Error during connection upgrader: ", err)
		return
	}

	connectionId := uuid.New().String()
	rawUid, ok := c.Get("uid")
	if !ok {
		log.Error("Error during getting uid from context")
		return
	}
	uid, ok := rawUid.(string)
	if !ok {
		log.Error("Error during casting uid to string")
		return
	}

	printStat(connectionId, uid, "connected")

	var socketBundle *UserSocketBundle
	var bundleExists bool
	if socketBundle, bundleExists = socketBundles[uid]; !bundleExists {
		// socket bundle has to be created
		socketBundle = NewUserSocketBundle(uid)
		socketBundles[uid] = socketBundle
	}

	socketBundle.AddSocket(connectionId, conn)

	defer func() {
		socketBundle.RemoveSocket("access token")
		conn.Close()
	}()

	conn.SetCloseHandler(func(code int, text string) error {
		printStat(connectionId, uid, fmt.Sprintf("Disconnected: %d", code))
		return nil
	})

	for {
		// read in a message
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Debug("Error during reading message: ", err)
			return
		}

		recvPacket, err := ToPacket(msg)
		if err != nil {
			// uncaught raw messages
			log.Warnf("Uncaught raw message: %s", string(msg))
			continue
		}

		printStat(
			connectionId,
			uid,
			fmt.Sprintf("[%s]message<%s> %v", shorten(recvPacket.RequestId), recvPacket.Topic, recvPacket.Data),
		)

		// find handler
		handler, ok := socketHandlers[recvPacket.Topic]
		if !ok {
			log.Warnf("Uncaught message: %s", string(msg))
			continue
		}

		// handle message
		resp, err := handler(conn, uid, recvPacket.Data)
		sendPacket := &SocketSendPacket{
			Topic:      recvPacket.Topic,
			Data:       resp,
			RequestId:  recvPacket.RequestId,
			Success:    err == nil,
			ErrMessage: "",
		}
		if err != nil {
			sendPacket.ErrMessage = err.Error()
		}

		sendMessage, err := sendPacket.bytes()
		if err != nil {
			log.Error("Error during creating packet: ", err)
			continue
		}

		// write out a message
		if err := conn.WriteMessage(msgType, sendMessage); err != nil {
			log.Error("Error during writing message: ", err)
			continue
		}
	}
}

func shorten(str string) string {
	if len(str) > 5 {
		return str[:5]
	}
	return str
}

func printStat(connectionId string, uid string, text string) {
	log.Infof("Client[%s] User[%s]: %s", shorten(connectionId), uid, text)
}

func UseSocketRouter(g *gin.RouterGroup) {
	sg := g.Group("/websocket")
	sg.Use(AuthMiddleware)
	sg.Any("/connect", SocketV1)
}
