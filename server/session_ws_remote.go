package server

import (
	"context"
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type sessionWSRemote struct {
	logger     *zap.Logger       `json:"-"`
	Id         uuid.UUID         `json:"id"`
	Fmt        SessionFormat     `json:"sessionFormat"`
	UId        uuid.UUID         `json:"userId"`
	UName      string            `json:"userName"`
	MapVars    map[string]string `json:"vars"`
	ExpUnix    int64             `json:"expire"`
	ClientIp   string            `json:"clientIp"`
	ClientPORT string            `json:"clientPort"`
	LANG       string            `json:"lang"`
	cmdEvent   CmdEvent
	NodeIp     string `json:"nodeIp"`
}

func NewSessionWSRemote() Session {
	ss := sessionWSRemote{}
	return &ss
}

func (sr *sessionWSRemote) Copy(s Session) Session {
	sr.Id = s.ID()
	sr.UId = s.UserID()
	sr.UName = s.Username()
	sr.MapVars = s.Vars()
	sr.ExpUnix = s.Expiry()
	sr.ClientIp = s.ClientIP()
	sr.ClientPORT = s.ClientPort()
	sr.LANG = s.Lang()
	return sr
}

func NewSessionWSRemoteFromJson(data []byte) Session {
	ss := sessionWSRemote{}
	json.Unmarshal(data, &ss)
	return &ss
}

func (s *sessionWSRemote) Logger() *zap.Logger {
	return s.logger
}
func (s *sessionWSRemote) ID() uuid.UUID {
	return s.Id
}
func (s *sessionWSRemote) UserID() uuid.UUID {
	return s.UId
}
func (s *sessionWSRemote) Vars() map[string]string {
	return s.MapVars
}
func (s *sessionWSRemote) ClientIP() string {
	return s.ClientIp
}
func (s *sessionWSRemote) ClientPort() string {
	return s.ClientPORT
}
func (s *sessionWSRemote) Lang() string {
	return s.LANG
}

func (s *sessionWSRemote) Context() context.Context {
	return context.Background()
}

func (s *sessionWSRemote) Username() string {
	return s.UName
}
func (s *sessionWSRemote) SetUsername(newUserName string) {
	// todo send cmd event
	s.cmdEvent.SendMessage(&CmdMessage{
		SessionId: s.ID(),
		TypeCmd:   CmdSessionSetUserName,
		Reliable:  false,
		Payload:   []byte(newUserName),
	})
}

func (s *sessionWSRemote) Expiry() int64 {
	return s.ExpUnix
}
func (s *sessionWSRemote) Consume() {
	// session remote do nothing
	s.logger.Info("Session remote not implemnt consume")
}

func (s *sessionWSRemote) Format() SessionFormat {
	return s.Fmt
}
func (s *sessionWSRemote) Send(envelope *rtapi.Envelope, reliable bool) error {
	// todo send cmd event
	if s.cmdEvent == nil {
		s.logger.Error("Session remote not implemnt send")
		return errors.New("not implement")
	}
	s.cmdEvent.SendMessage(&CmdMessage{
		SessionId: s.ID(),
		TypeCmd:   CmdSend,
		Payload:   []byte(envelope.String()),
		Reliable:  reliable,
	})
	return nil
}
func (s *sessionWSRemote) SendBytes(payload []byte, reliable bool) error {
	if s.cmdEvent == nil {
		s.logger.Error("Session remote not implemnt send")
		return errors.New("not implement")
	}
	s.cmdEvent.SendMessage(&CmdMessage{
		SessionId: s.ID(),
		TypeCmd:   CmdSendBytes,
		Payload:   payload,
		Reliable:  reliable,
	})
	return nil
}

func (s *sessionWSRemote) Close(msg string, reason runtime.PresenceReason, envelopes ...*rtapi.Envelope) {
	if s.cmdEvent == nil {
		s.logger.Error("Session remote not implemnt send")
		return
	}
	payloadStruct := SessionClosePayload{
		Msg:    msg,
		Reason: uint8(reason),
	}
	for _, envelope := range envelopes {
		payloadStruct.Envelopes = append(payloadStruct.Envelopes, envelope.String())
	}
	payload, _ := json.Marshal(payloadStruct)
	s.cmdEvent.SendMessage(&CmdMessage{
		SessionId: s.ID(),
		TypeCmd:   CmdSendBytes,
		Payload:   payload,
		Reliable:  false,
	})
}
