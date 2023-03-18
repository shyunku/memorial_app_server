package v1

import (
	"github.com/gorilla/websocket"
	"memorial_app_server/log"
	"net/http"
)

func UseSocketV1(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("Error during connection upgrader: ", err)
		return
	}
	defer conn.Close()

	log.Infof("Client connected: %s", conn.RemoteAddr().String())

	for {
		// read in a message
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			log.Error("Error during reading message: ", err)
			return
		}
		log.Debug("Message received: ", string(msg))

		// write out a message
		if err := conn.WriteMessage(msgType, msg); err != nil {
			log.Error("Error during writing message: ", err)
			return
		}
	}
}
