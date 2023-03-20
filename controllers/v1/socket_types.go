package v1

import (
	"encoding/json"
	"github.com/gorilla/websocket"
)

type SocketPacket struct {
	Topic     string      `json:"topic"`
	Data      interface{} `json:"data"`
	RequestId string      `json:"reqId"`
}

func ToPacket(raw []byte) (*SocketPacket, error) {
	var packet SocketPacket
	err := json.Unmarshal(raw, &packet)
	if err != nil {
		return nil, err
	}
	return &packet, nil
}

type SocketSendPacket struct {
	Topic      string      `json:"topic"`
	Data       interface{} `json:"data"`
	RequestId  string      `json:"reqId"`
	Success    bool        `json:"success"`
	ErrMessage string      `json:"err_message"`
}

func (p *SocketSendPacket) bytes() ([]byte, error) {
	return json.Marshal(p)
}

type UserSocket struct {
	ConnectionId string
	Conn         *websocket.Conn
}

func NewUserSocket(connectionId string, conn *websocket.Conn) *UserSocket {
	return &UserSocket{
		ConnectionId: connectionId,
		Conn:         conn,
	}
}

type UserSocketBundle struct {
	UserId  string
	sockets map[string]*UserSocket
}

func NewUserSocketBundle(userId string) *UserSocketBundle {
	return &UserSocketBundle{
		UserId:  userId,
		sockets: map[string]*UserSocket{},
	}
}

func (b *UserSocketBundle) AddSocket(connectionId string, conn *websocket.Conn) {
	b.sockets[connectionId] = NewUserSocket(connectionId, conn)
}

func (b *UserSocketBundle) RemoveSocket(connectionId string) {
	delete(b.sockets, connectionId)
}

func (b *UserSocketBundle) GetSocket(connectionId string) *UserSocket {
	return b.sockets[connectionId]
}

/* -------------------------------- Custom -------------------------------- */

type TxSocketRequest struct {
	From      string `json:"from"`
	Type      int64  `json:"type"`
	Timestamp int64  `json:"timestamp"`
	Content   []byte `json:"content"`

	TargetBlockNumber string // target block number as string (e.g. "1051240")
}
