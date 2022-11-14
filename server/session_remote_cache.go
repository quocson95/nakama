package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
)

type SessionRemoteCache interface {
	SessionCache
	Get() CustomSessionCache
}

type SessionRemoteCacheUser struct {
	// uid         uuid.UUID
	ctx         context.Context
	ctxCancelFn context.CancelFunc
	rdb         *redis.Client
	config      Config
}

const DestRemoteHost = "http://103.226.250.195:7350"

func NewRemoteSessionCache(rdb *redis.Client) SessionCache {
	ctx, ctxCancelFn := context.WithCancel(context.Background())

	s := SessionRemoteCacheUser{
		ctx:         ctx,
		ctxCancelFn: ctxCancelFn,
		rdb:         rdb,
	}
	return &s
}

const KeySessionFmt = "session:%s:%s"
const KeySessionRefreshFmt = "session-refresh:%s:%s"

type CustomSessionCache struct {
	NodeAddress string
}

func ParseCustomSessionId(data []byte) *CustomSessionCache {
	c := CustomSessionCache{}
	err := json.Unmarshal(data, &c)
	if err != nil {
		return nil
	}
	return &c
}
func (s *SessionRemoteCacheUser) getSessionTokens(userID uuid.UUID, token string) (*CustomSessionCache, error) {
	key := fmt.Sprintf(KeySessionFmt, userID.String(), token)
	sid, err := s.rdb.Get(s.ctx, key).Result()
	return ParseCustomSessionId([]byte(sid)), err
}

func (s *SessionRemoteCacheUser) getRefreshTokens(userID uuid.UUID, token string) (*CustomSessionCache, error) {
	key := fmt.Sprintf(KeySessionRefreshFmt, userID.String(), token)
	sid, err := s.rdb.Get(s.ctx, key).Result()
	return ParseCustomSessionId([]byte(sid)), err
}

func (s *SessionRemoteCacheUser) Stop() {
	s.ctxCancelFn()
}

func (s *SessionRemoteCacheUser) IsValidSession(userID uuid.UUID, exp int64, token string) bool {
	customSession, err := s.getSessionTokens(userID, token)
	if err != nil || customSession == nil {
		return false
	}
	return true
}

func (s *SessionRemoteCacheUser) IsValidRefresh(userID uuid.UUID, exp int64, token string) bool {
	customSession, err := s.getRefreshTokens(userID, token)
	if err != nil || customSession == nil {
		return false
	}
	return true
}

func (s *SessionRemoteCacheUser) Add(userID uuid.UUID, sessionExp int64, sessionToken string, refreshExp int64, refreshToken string) {

	customSession := CustomSessionCache{
		NodeAddress: DestRemoteHost,
	}
	customSessionData, _ := json.Marshal(customSession)
	// save session token
	{
		key := fmt.Sprintf(KeySessionFmt, userID.String(), sessionToken)
		s.rdb.Set(s.ctx, key, customSessionData, time.Duration(sessionExp))
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionRefreshFmt, userID.String(), refreshToken)
		s.rdb.Set(s.ctx, key, customSessionData, time.Duration(refreshExp))
	}
}

func (s *SessionRemoteCacheUser) Get(userID uuid.UUID, sessionToken string) *CustomSessionCache {
	key := fmt.Sprintf(KeySessionFmt, userID.String(), sessionToken)
	data, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		return nil
	}
	c := CustomSessionCache{}
	json.Unmarshal([]byte(data), &c)
	return &c
}

func (s *SessionRemoteCacheUser) Remove(userID uuid.UUID, sessionExp int64, sessionToken string, refreshExp int64, refreshToken string) {
	// remove session token
	{
		key := fmt.Sprintf(KeySessionFmt, userID.String(), sessionToken)
		s.rdb.Del(s.ctx, key)
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionRefreshFmt, userID.String(), refreshToken)
		s.rdb.Del(s.ctx, key)
	}
}

func (s *SessionRemoteCacheUser) RemoveAll(userID uuid.UUID) {
	// remove session token
	{
		key := fmt.Sprintf(KeySessionFmt, userID.String(), "*")
		s.rdb.Del(s.ctx, key)
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionRefreshFmt, userID.String(), "*")
		s.rdb.Del(s.ctx, key)
	}
}

func (s *SessionRemoteCacheUser) Ban(userIDs []uuid.UUID) {

	for _, userID := range userIDs {
		s.RemoveAll(userID)
	}
}

func (s *SessionRemoteCacheUser) Unban(userIDs []uuid.UUID) {}
