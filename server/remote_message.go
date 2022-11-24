package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

type TypeCmdMessage int

const (
	CmdRemoveSession          TypeCmdMessage = 1
	CmdSessionSetUserName     TypeCmdMessage = 2
	CmdSessionRegSingleSesion TypeCmdMessage = 3
	CmdSessionRegDisconnect   TypeCmdMessage = 4
	CmdSend                   TypeCmdMessage = 5
	CmdSendBytes              TypeCmdMessage = 6
)

type CmdMessage struct {
	SessionId uuid.UUID      `json:"sessionId"`
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
	SendMessage(msg *CmdMessage) error
	// RecvMessage(msg *CmdMessage)
}
type RemoteCmdEvent struct {
	config Config
	logger *zap.Logger
	rdb    *redis.Client
}

func NewRemoteCmdEvent(config Config, logger *zap.Logger, rdb *redis.Client) CmdEvent {
	cmd := &RemoteCmdEvent{
		logger: logger,
		rdb:    rdb,
		config: config,
	}
	return cmd
}

func (r *RemoteCmdEvent) SendMessage(msg *CmdMessage) error {
	// todo
	data, _ := json.Marshal(msg)
	body := bytes.NewBuffer(data)
	req, err := http.NewRequest("POST",
		fmt.Sprintf("http://%s/v2/proxy-node", msg.NodeIp),
		body)
	query := req.URL.Query()
	query.Add("http_key", r.config.GetRuntime().HTTPKey)
	req.URL.RawQuery = query.Encode()
	defer func() {
		if err != nil {
			r.logger.
				With(zap.Error(err)).
				With(zap.ByteString("body", data)).
				Error("Send msg to node failed")
		} else {
			r.logger.
				With(zap.ByteString("body", data)).
				Info("Send msg to node success")
		}
	}()
	if err != nil {
		return err
	}
	var resp *http.Response
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		var body []byte
		if resp.Body != nil {
			body, _ = io.ReadAll(resp.Body)
			resp.Body.Close()
		}
		r.logger.
			With(zap.Int("status", resp.StatusCode)).
			With(zap.ByteString("response body", body)).
			Error("Send msg to node failed")
	}
	return nil
}

func (r *RemoteCmdEvent) RecvMessage(msg *CmdMessage) {}
