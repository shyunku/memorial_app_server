package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"memorial_app_server/log"
	"memorial_app_server/service/state"
	"memorial_app_server/util"
)

type socketHandler func(socket *UserSocket, uid string, data interface{}) (interface{}, error)

var (
	socketHandlers = map[string]socketHandler{
		"test":               test,
		"transaction":        handleTransaction,
		"waitingBlockNumber": waitingBlockNumber,
		"syncBlocks":         syncBlocks,
		"commitTransactions": commitTransactions,
	}
	socketBundles = map[string]*UserSocketBundle{}
)

func test(socket *UserSocket, uid string, data interface{}) (interface{}, error) {
	// data to string
	str, ok := data.(string)
	if !ok {
		return nil, errors.New("invalid data type")
	}
	return str, nil
}

func handleTransaction(socket *UserSocket, uid string, data interface{}) (interface{}, error) {
	var request TxSocketRequest
	if err := util.InterfaceToStruct(data, &request); err != nil {
		log.Errorf("Failed to unmarshal data: %v", data)
		return nil, fmt.Errorf("invalid request: check format")
	}
	userChain := state.Chains.GetChain(uid)

	log.Debug(request)

	// check if targetBlockNumber is valid
	waitingBlockNumber := userChain.GetWaitingBlockNumber()
	if request.TargetBlockNumber != waitingBlockNumber {
		// different block number
		log.Errorf("Invalid target block number: waiting for block #%d, but #%d given", waitingBlockNumber, request.TargetBlockNumber)
		return nil, fmt.Errorf("invalid block number: waiting for block #%d", waitingBlockNumber)
	}

	// check if transaction is valid
	tx := state.NewTransaction(uid, request.Type, request.Timestamp, request.Content)
	if err := tx.Validate(); err != nil {
		log.Error("Invalid transaction")
		return nil, fmt.Errorf("invalid request: %s", err.Error())
	}

	// apply transaction
	newBlock, err := userChain.ApplyTransaction(tx)
	if err != nil {
		log.Errorf("Error during applying transaction: %v", err)
		return nil, fmt.Errorf("failed to apply transaction: %s", err.Error())
	}

	go func() {
		// broadcast transaction to same user connections
		bundle, ok := socketBundles[uid]
		if !ok {
			log.Warnf("Couldn't find socket bundle for user %s", uid)
		}

		for _, sock := range bundle.sockets {
			// except sender
			if sock.ConnectionId == socket.ConnectionId {
				continue
			}

			// send transaction
			if err := sock.Emitter("broadcast_transaction", newBlock); err != nil {
				log.Warnf("Failed to broadcast transaction to user %s [%s]", uid, sock.ConnectionId)
			}
		}
	}()

	return nil, nil
}

func waitingBlockNumber(socket *UserSocket, uid string, data interface{}) (interface{}, error) {
	userChain := state.Chains.GetChain(uid)
	return userChain.GetWaitingBlockNumber(), nil
}

func syncBlocks(socket *UserSocket, uid string, data interface{}) (interface{}, error) {
	var request SyncBlocksSocketRequest
	if err := util.InterfaceToStruct(data, &request); err != nil {
		log.Errorf("Failed to unmarshal data: %v", data)
		return nil, fmt.Errorf("invalid request: check format")
	}

	userChain := state.Chains.GetChain(uid)

	startBlockNumber := request.StartBlockNumber
	endBlockNumber := request.EndBlockNumber

	// check if startBlockNumber is smaller than endBlockNumber
	if startBlockNumber > endBlockNumber {
		log.Error("Invalid block number range")
		return nil, fmt.Errorf("invalid block number range: start block number is greater than end block number")
	}

	// fetch blocks from chain
	blocks, err := userChain.GetBlocksByInterval(startBlockNumber, endBlockNumber)
	if err != nil {
		log.Error("Error during fetching blocks")
		return nil, fmt.Errorf("failed to fetch blocks: %s", err.Error())
	}

	return blocks, nil
}

func commitTransactions(socket *UserSocket, uid string, data interface{}) (interface{}, error) {
	request, ok := data.(*CommitTxBundleSocketRequest)
	if !ok {
		log.Error("Invalid data type")
		return nil, fmt.Errorf("invalid request: check format")
	}

	for _, txReq := range *request {
		_, err := handleTransaction(socket, uid, txReq)
		if err != nil {
			log.Error("Error during handling transaction: ", err)
			return nil, err
		}
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
		log.Errorf("Error during casting uid to string: %v", rawUid)
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

	socket := socketBundle.AddSocket(connectionId, conn, func(topic string, data interface{}) error {
		sendPacket := &SocketSendPacket{
			Topic:      topic,
			Data:       data,
			RequestId:  "",
			Success:    true,
			ErrMessage: "",
		}

		sendMessage, err := sendPacket.bytes()
		if err != nil {
			log.Error("Error during creating packet: ", err)
			return err
		}

		// write out a message
		if err := conn.WriteMessage(websocket.BinaryMessage, sendMessage); err != nil {
			log.Error("Error during writing message: ", err)
			return err
		}

		return nil
	})

	defer func() {
		socketBundle.RemoveSocket(connectionId)
		if socketBundle.GetSize() == 0 {
			delete(socketBundles, uid)
		}
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
			fmt.Sprintf("[%s] message<%s> %v", shorten(recvPacket.RequestId), recvPacket.Topic, recvPacket.Data),
		)

		// find handler
		handler, ok := socketHandlers[recvPacket.Topic]
		if !ok {
			log.Warnf("Uncaught message: %s", string(msg))
			continue
		}

		// handle message
		resp, err := handler(socket, uid, recvPacket.Data)
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
