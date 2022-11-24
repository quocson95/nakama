package server

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type RemoteSessionRegistry struct {
	LocalSessionRegistry
	cmdEvent           CmdEvent
	rdb                *redis.Client
	ctx                context.Context
	nodeIp             string
	logger             *zap.Logger
	protojsonMarshaler *protojson.MarshalOptions
}

func NewRemoteSessionRegistry(metrics Metrics, cmdEvent CmdEvent, rdb *redis.Client, logger *zap.Logger, protojsonMarshaler *protojson.MarshalOptions, nodeIp string) SessionRegistry {
	r := &RemoteSessionRegistry{
		LocalSessionRegistry: LocalSessionRegistry{
			metrics:      metrics,
			sessions:     &sync.Map{},
			sessionCount: atomic.NewInt32(0),
		},
	}
	r.cmdEvent = cmdEvent
	r.rdb = rdb
	r.nodeIp = nodeIp
	r.ctx = context.Background()
	r.logger = logger
	r.protojsonMarshaler = protojsonMarshaler
	return r
}

func (r *RemoteSessionRegistry) Stop() {}

func (r *RemoteSessionRegistry) Count() int {
	key := KeyHsetSessionRegFmt
	count, _ := r.rdb.HLen(r.ctx, key).Result()
	return int(count)
}

func (r *RemoteSessionRegistry) Get(sessionID uuid.UUID) Session {
	session := r.LocalSessionRegistry.Get(sessionID)
	if session != nil {
		return session
	}
	// get from redis for session in another node
	key := KeyHsetSessionRegFmt
	data, err := r.rdb.HGet(r.ctx, key, sessionID.String()).Result()
	if err != nil {
		// not found, maybe sessionID invalid
		r.logger.With(zap.String("sid", sessionID.String())).
			Error("get session registry failed")
		return nil
	}
	// remoteSession := NewSessionWSRemoteFromJson([]byte(data))
	remoteSession := NewSessionWSRemote(r.logger, r.cmdEvent, r.protojsonMarshaler).(*sessionWSRemote)
	remoteSession.FromJson([]byte(data))
	// session not found in local mem
	// but found on redis, nodeip in redis = this node ip
	// => session socket maybe invalid -> remove and return as not found
	//
	if remoteSession.NodeIp == r.nodeIp {
		r.logger.With(zap.String("sid", sessionID.String())).
			Error("Session found on redis, but not found in any node")
		r.rdb.HDel(r.ctx, key, sessionID.String())
		return nil
	}
	// save to local

	r.sessions.Store(remoteSession.ID(), remoteSession)
	count := r.Count()
	r.metrics.GaugeSessions(float64(count))
	return remoteSession
}

func (r *RemoteSessionRegistry) Add(session Session) {
	r.sessions.Store(session.ID(), session)
	count := r.Count()
	r.metrics.GaugeSessions(float64(count))
	// save info to redis for another node
	remoteSession := NewSessionWSRemote(r.logger, r.cmdEvent, r.protojsonMarshaler).(*sessionWSRemote)
	remoteSession.Copy(session)
	remoteSession.NodeIp = r.nodeIp

	data, err := json.Marshal(remoteSession)
	if err == nil {
		r.rdb.HSet(r.ctx, KeyHsetSessionRegFmt, session.ID().String(), data)
	}
}

func (r *RemoteSessionRegistry) Remove(sessionID uuid.UUID) {
	session := r.LocalSessionRegistry.Get(sessionID)
	if session != nil {
		// session in this node
		r.sessions.Delete(sessionID)
		return
	}
	// session in another node

	key := KeyHsetSessionRegFmt
	data, err := r.rdb.HGet(r.ctx, key, sessionID.String()).Result()
	if err != nil {
		// not found
		return
	}
	_, err = r.rdb.HDel(r.ctx, key, sessionID.String()).Result()
	if err != nil {
		r.logger.With(zap.String("sid", sessionID.String())).
			With(zap.Error(err)).
			Error("Session remove on redis failed")
	}
	// todo broadcast to all node
	remoteSession := NewSessionWSRemote(r.logger, r.cmdEvent, r.protojsonMarshaler).(*sessionWSRemote)
	remoteSession.FromJson([]byte(data))
	r.cmdEvent.SendMessage(&CmdMessage{
		NodeIp:   remoteSession.NodeIp,
		TypeCmd:  CmdRemoveSession,
		Reliable: false,
		Payload:  sessionID.Bytes(),
	})
}

func (r *RemoteSessionRegistry) Disconnect(ctx context.Context, sessionID uuid.UUID, reason ...runtime.PresenceReason) error {
	session := r.LocalSessionRegistry.Get(sessionID)
	if session == nil {
		return nil
	}
	remoteSession, isRemoteSession := session.(*sessionWSRemote)
	if !isRemoteSession && session != nil {
		// session in this node
		r.LocalSessionRegistry.Disconnect(ctx, sessionID, reason...)
		return nil
	}
	// session in another node
	payloadStruct := SessionDisconnectPayload{
		Msg:       "",
		SessionId: []string{sessionID.String()},
	}
	for _, r := range reason {
		payloadStruct.Reasons = append(payloadStruct.Reasons, uint8(r))
	}
	cmdMsg := &CmdMessage{
		NodeIp:   remoteSession.NodeIp,
		TypeCmd:  CmdSessionRegDisconnect,
		Reliable: false,
	}
	cmdMsg.Payload, _ = json.Marshal(payloadStruct)
	r.cmdEvent.SendMessage(cmdMsg)
	return nil
}

func (r *RemoteSessionRegistry) SingleSession(ctx context.Context, tracker Tracker, userID, sessionID uuid.UUID) {
	session := r.LocalSessionRegistry.Get(sessionID)
	if session == nil {
		return
	}
	remoteSession, isRemoteSession := session.(*sessionWSRemote)
	if !isRemoteSession {
		// session in this node
		r.LocalSessionRegistry.SingleSession(ctx, tracker, userID, sessionID)
		return
	}
	// session in another node
	cmdMsg := &CmdMessage{
		NodeIp:   remoteSession.NodeIp,
		TypeCmd:  CmdSessionRegSingleSesion,
		Reliable: false,
		Payload:  userID.Bytes(),
	}
	r.cmdEvent.SendMessage(cmdMsg)
}
