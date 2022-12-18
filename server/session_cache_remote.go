package server

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofrs/uuid"
	"go.uber.org/zap"
)

const lenTokenPrint = 10

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
	logger      *zap.Logger
}

func (s *SessionRemoteCacheUser) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func NewRemoteSessionCache(rdb *redis.Client, logger *zap.Logger, conf Config) SessionCache {
	ctx, ctxCancelFn := context.WithCancel(context.Background())

	s := SessionRemoteCacheUser{
		ctx:         ctx,
		ctxCancelFn: ctxCancelFn,
		rdb:         rdb,
		logger:      logger,
		config:      conf,
	}
	return &s
}

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
	if err != nil {
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).With(zap.Error(err)).Info("get session cache failure")
		return nil, err
	}
	c := ParseCustomSessionId([]byte(sid))
	s.logger.
		With(zap.String("key", shortString(key, lenTokenPrint))).
		With(zap.String("node ip", c.NodeAddress)).
		Info("get session cache successful")

	return c, err
}

func (s *SessionRemoteCacheUser) getRefreshTokens(userID uuid.UUID, token string) (*CustomSessionCache, error) {
	key := fmt.Sprintf(KeySessionCacheRefreshFmt, userID.String(), token)
	sid, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		s.logger.
			With(zap.String("key", shortString(key, lenTokenPrint))).
			With(zap.Error(err)).
			Info("get session refresh failure")
		return nil, err
	}
	c := ParseCustomSessionId([]byte(sid))
	s.logger.
		With(zap.String("key", shortString(key, lenTokenPrint))).
		With(zap.String("node ip", c.NodeAddress)).
		Info("get session refresh successful")

	return c, err
}

func (s *SessionRemoteCacheUser) Stop() {
	s.ctxCancelFn()
}

func (s *SessionRemoteCacheUser) IsValidSession(userID uuid.UUID, exp int64, token string) bool {
	customSession, err := s.getSessionTokens(userID, token)
	if err != nil || customSession == nil {
		s.logger.With(zap.Error(err)).Error("get cache failed")
		return false
	}
	return true
}

func (s *SessionRemoteCacheUser) IsValidRefresh(userID uuid.UUID, exp int64, token string) bool {
	customSession, err := s.getRefreshTokens(userID, token)
	if err != nil || customSession == nil {
		s.logger.With(zap.Error(err)).Error("get cache failed")
		return false
	}
	return true
}

func (s *SessionRemoteCacheUser) Add(userID uuid.UUID, sessionExp int64, sessionToken string, refreshExp int64, refreshToken string) {

	// host := fmt.Sprintf("%s:%d", s.config.GetPublicIP(), s.config.GetSocket().Port)
	host := s.config.GetPublicIP()
	customSession := CustomSessionCache{
		NodeAddress: host,
	}
	customSessionData, _ := json.Marshal(customSession)
	// save session token
	{
		key := fmt.Sprintf(KeySessionFmt, userID.String(), sessionToken)
		_, err := s.rdb.Set(s.ctx, key, customSessionData, time.Duration(sessionExp-time.Now().Unix())*time.Second).Result()
		if err == nil {
			s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).Info("add session cache successful")
		} else {
			s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).With(zap.Error(err)).Info("add session cache failure")
		}
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionCacheRefreshFmt, userID.String(), refreshToken)
		s.rdb.Set(s.ctx, key, customSessionData, time.Duration(refreshExp-time.Now().Unix())*time.Second).Result()
	}
}

func (s *SessionRemoteCacheUser) Get(userID uuid.UUID, sessionToken string) *CustomSessionCache {
	key := fmt.Sprintf(KeySessionFmt, userID.String(), sessionToken)
	data, err := s.rdb.Get(s.ctx, key).Result()
	if err != nil {
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).With(zap.Error(err)).Error("Get failed")
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
		_, err := s.rdb.Del(s.ctx, key).Result()
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).
			With(zap.Error(err)).Info("Remove session key")
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionCacheRefreshFmt, userID.String(), refreshToken)
		_, err := s.rdb.Del(s.ctx, key).Result()
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).
			With(zap.Error(err)).Info("Remove refresh token key")
	}
}

func (s *SessionRemoteCacheUser) RemoveAll(userID uuid.UUID) {
	// remove session token
	{
		key := fmt.Sprintf(KeySessionFmt, userID.String(), "*")
		_, err := s.rdb.Del(s.ctx, key).Result()
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).
			With(zap.Error(err)).Info("Remove all session user")
	}
	//// save session refresh token
	{
		key := fmt.Sprintf(KeySessionCacheRefreshFmt, userID.String(), "*")
		_, err := s.rdb.Del(s.ctx, key).Result()
		s.logger.With(zap.String("key", shortString(key, lenTokenPrint))).
			With(zap.Error(err)).Info("Remove all refresh token user")
	}
}

func (s *SessionRemoteCacheUser) Ban(userIDs []uuid.UUID) {

	for _, userID := range userIDs {
		s.RemoveAll(userID)
	}
}

func (s *SessionRemoteCacheUser) Unban(userIDs []uuid.UUID) {}

func shortString(str string, length int) string {
	lenStr := len(str)
	if lenStr <= length {
		return str
	}
	return str[lenStr-length : lenStr]
}
