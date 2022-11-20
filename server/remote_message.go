package server

import (
	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
)

type TypeCmdMessage int

const (
	CmdRemoveSession TypeCmdMessage = iota
	CmdSessionSetUserName
	CmdSessionRegSingleSesion
	CmdSessionRegDisconnect
	CmdSend
	CmdSendBytes
)

type CmdMessage struct {
	SessionId uuid.UUID      `json:"-"`
	NodeIp    string         `json:"nodeIP"`
	TypeCmd   TypeCmdMessage `json:"typeCmd"`
	Reliable  bool           `json:"reliable"`
	Payload   []byte         `json:"payload"`
}

type SessionClosePayload struct {
	Msg       string   `json:"msg"`
	Reason    uint8    `json:"reason"`
	Envelopes []string `json:"envelopes"`
}

type SessionDisconnectPayload struct {
	Msg       string   `json:"msg"`
	Reasons   []uint8  `json:"reasons"`
	SessionId []string `json:"sessionId"`
}

type CmdEvent interface {
	SendMessage(msg *CmdMessage)
	// RecvMessage(msg *CmdMessage)
}
type RemoteCmdEvent struct {
	rdb *redis.Client
}

func NewRemoteCmdEvent(rdb *redis.Client) CmdEvent {
	cmd := &RemoteCmdEvent{
		rdb: rdb,
	}
	return cmd
}

func (r *RemoteCmdEvent) SendMessage(msg *CmdMessage) {
	// todo
}

func (r *RemoteCmdEvent) RecvMessage(msg *CmdMessage) {}
