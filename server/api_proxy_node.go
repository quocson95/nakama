package server

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"github.com/heroiclabs/nakama-common/rtapi"
	"github.com/heroiclabs/nakama-common/runtime"
	"go.uber.org/zap"
)

// import (
// 	"context"
// 	"encoding/json"
// 	"io/ioutil"
// 	"net/http"
// 	"strings"
// 	"time"

// 	"github.com/gofrs/uuid"
// 	grpcgw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
// 	"github.com/heroiclabs/nakama-common/rtapi"
// 	"github.com/heroiclabs/nakama-common/runtime"
// 	"go.uber.org/zap"
// 	"google.golang.org/grpc/codes"
// )

// func NewProxyNodeHandler(s *ApiServer,
// 	sessionRegistry SessionRegistry,
// 	tracker Tracker) func(w http.ResponseWriter, r *http.Request) {
// 	fnn := func(ctx context.Context, headers, queryParams map[string][]string, userID, username string, vars map[string]string, expiry int64, sessionID, clientIP, clientPort, lang, payload string) (string, error, codes.Code) {
// 		var cmdMsg CmdMessage
// 		err := json.Unmarshal([]byte(payload), &cmdMsg)
// 		if err != nil {
// 			s.logger.
// 				With(zap.Error(err)).
// 				Error("Parse payload failed")
// 			return "", err, codes.Internal
// 		}
// 		// if cmdMsg.SessionId.String() == "" {
// 		// 	return "", errors.New("Missing session id"), codes.InvalidArgument
// 		// }
// 		s.logger.
// 			With(zap.String("session", cmdMsg.SessionId.String())).
// 			With(zap.Int("type cmd", int(cmdMsg.TypeCmd))).
// 			Debug("Exec cmd remote sesion")
// 		switch cmdMsg.TypeCmd {
// 		case CmdRemoveSession:
// 			sessionRegistry.Remove(cmdMsg.SessionId)
// 		case CmdSessionSetUserName:
// 			// todo
// 		case CmdSessionRegSingleSesion:
// 			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 			defer cancel()
// 			userID, _ := uuid.FromString(string(cmdMsg.Payload))
// 			sessionRegistry.SingleSession(ctx,
// 				tracker, userID, cmdMsg.SessionId)
// 		case CmdSessionRegDisconnect:
// 			sessionDisStruct := SessionDisconnectPayload{}
// 			err = json.Unmarshal(cmdMsg.Payload, &sessionDisStruct)
// 			if err != nil {
// 				s.logger.
// 					With(zap.Error(err)).
// 					Error("Parse payload CmdSessionRegDisconnect")
// 				return "", err, codes.Internal
// 			}
// 			listSessionUuid := make([]uuid.UUID, 0, len(sessionDisStruct.SessionId))
// 			for _, str := range sessionDisStruct.SessionId {
// 				sessionUUID, _ := uuid.FromString(str)
// 				listSessionUuid = append(listSessionUuid, sessionUUID)
// 			}
// 			listReason := make([]runtime.PresenceReason, 0,
// 				len(sessionDisStruct.Reasons))
// 			for _, r := range sessionDisStruct.Reasons {
// 				listReason = append(listReason, runtime.PresenceReason(r))
// 			}
// 			for _, uuid := range listSessionUuid {
// 				{
// 					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 					defer cancel()
// 					sessionRegistry.Disconnect(ctx, uuid, listReason...)
// 				}
// 			}
// 		case CmdSend:
// 			session := sessionRegistry.Get(cmdMsg.SessionId)
// 			if session == nil {
// 				s.logger.
// 					With(zap.String("session id", cmdMsg.SessionId.String())).
// 					Error("Session is nil")
// 				return "", err, codes.Internal
// 			}
// 			envelope := &rtapi.Envelope{}
// 			err = json.Unmarshal(cmdMsg.Payload, envelope)
// 			if err != nil {
// 				s.logger.
// 					With(zap.Error(err)).
// 					With(zap.ByteString("payload", cmdMsg.Payload)).
// 					Error("Parse payload rtapi.Envelope failed")
// 				return "", err, codes.Internal
// 			}
// 			err = session.Send(envelope, cmdMsg.Reliable)
// 			if err != nil {
// 				s.logger.
// 					With(zap.String("session id", cmdMsg.SessionId.String())).
// 					With(zap.ByteString("payload", cmdMsg.Payload)).
// 					With(zap.Error(err)).
// 					Error("Session Send failed")
// 				return "", err, codes.Internal
// 			}
// 		case CmdSendBytes:
// 			session := sessionRegistry.Get(cmdMsg.SessionId)
// 			if session == nil {
// 				s.logger.
// 					With(zap.String("session id", cmdMsg.SessionId.String())).
// 					Error("Session is nil")
// 				return "", err, codes.Internal
// 			}
// 			err = session.SendBytes(cmdMsg.Payload, cmdMsg.Reliable)
// 			if err != nil {
// 				s.logger.
// 					With(zap.String("session id", cmdMsg.SessionId.String())).
// 					With(zap.ByteString("payload", cmdMsg.Payload)).
// 					With(zap.Error(err)).
// 					Error("Session Send failed")
// 				return "", err, codes.Internal
// 			}
// 		}
// 		return "", nil, codes.OK
// 	}
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		queryParams := r.URL.Query()
// 		httpKey := queryParams.Get("http_key")
// 		w.Header().Set("content-type", "application/json")
// 		if httpKey == "" || httpKey != s.config.GetRuntime().HTTPKey {
// 			w.WriteHeader(http.StatusUnauthorized)
// 			_, err := w.Write(httpKeyInvalidBytes)
// 			if err != nil {
// 				s.logger.Debug("Error writing response to client", zap.Error(err))
// 			}
// 			return
// 		}

// 		start := time.Now()
// 		var success bool
// 		var recvBytes, sentBytes int
// 		var err error
// 		id := strings.ToLower("proxy-node")

// 		// // After this point the RPC will be captured in metrics.
// 		defer func() {
// 			s.metrics.ApiRpc(id, time.Since(start), int64(recvBytes), int64(sentBytes), !success)
// 		}()
// 		if r.Method != "POST" {
// 			w.Header().Set("content-type", "application/json")
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
// 		// id = strings.ToLower("proxy-node")
// 		// fn := s.runtime.Rpc(id)
// 		// if fn == nil {
// 		// 	// No function registered for this ID.
// 		// 	w.Header().Set("content-type", "application/json")
// 		// 	w.WriteHeader(http.StatusNotFound)
// 		// 	sentBytes, err = w.Write(rpcFunctionNotFoundBytes)
// 		// 	if err != nil {
// 		// 		s.logger.Debug("Error writing response to client", zap.Error(err))
// 		// 	}
// 		// 	return
// 		// }
// 		var payload string
// 		b, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			if err.Error() == "http: request body too large" {
// 				w.Header().Set("content-type", "application/json")
// 				w.WriteHeader(http.StatusBadRequest)
// 				sentBytes, err = w.Write(requestBodyTooLargeBytes)
// 				if err != nil {
// 					s.logger.Debug("Error writing response to client", zap.Error(err))
// 				}
// 				return
// 			}

// 			// Other error reading request body.
// 			w.Header().Set("content-type", "application/json")
// 			w.WriteHeader(http.StatusInternalServerError)
// 			sentBytes, err = w.Write(internalServerErrorBytes)
// 			if err != nil {
// 				s.logger.Debug("Error writing response to client", zap.Error(err))
// 			}
// 			return
// 		}
// 		recvBytes = len(b)
// 		payload = string(b)
// 		queryParams.Del("http_key")
// 		uid := ""

// 		clientIP, clientPort := extractClientAddressFromRequest(s.logger, r)

// 		// Extract http headers
// 		headers := make(map[string][]string)
// 		for k, v := range r.Header {
// 			if k == "Grpc-Timeout" {
// 				continue
// 			}
// 			headers[k] = make([]string, 0, len(v))
// 			for _, h := range v {
// 				headers[k] = append(headers[k], h)
// 			}
// 		}

// 		var username string
// 		var vars map[string]string
// 		var expiry int64
// 		// Execute the function.
// 		result, fnErr, code := fnn(r.Context(), headers, queryParams, uid, username, vars, expiry, "", clientIP, clientPort, "", payload)
// 		if fnErr != nil {
// 			response, _ := json.Marshal(map[string]interface{}{"error": fnErr, "message": fnErr.Error(), "code": code})
// 			w.Header().Set("content-type", "application/json")
// 			w.WriteHeader(grpcgw.HTTPStatusFromCode(code))
// 			sentBytes, err = w.Write(response)
// 			if err != nil {
// 				s.logger.Debug("Error writing response to client", zap.Error(err))
// 			}
// 			return
// 		}
// 		response := []byte(result)
// 		if contentType := r.Header["Content-Type"]; len(contentType) > 0 {
// 			// Assume the request input content type is the same as the expected response.
// 			w.Header().Set("content-type", contentType[0])
// 		} else {
// 			// Don't know payload content-type.
// 			w.Header().Set("content-type", "text/plain")
// 		}
// 		w.WriteHeader(http.StatusOK)
// 		sentBytes, err = w.Write(response)
// 		if err != nil {
// 			s.logger.Debug("Error writing response to client", zap.Error(err))
// 			return
// 		}
// 		success = true
// 	}
// }

func RegisterPubSubHandlerSession(pubsub PubSubEvent, sessionRegistry SessionRegistry, tracker Tracker, logger *zap.Logger) {
	fn := func(data PubSubData) {
		handlerPubSubEvent(data, sessionRegistry, tracker, logger)
	}
	arr := make([]TypeData, 0)
	arr = append(arr,
		TypeDataRemoveSession,
		TypeDataSessionSetUserName,
		TypeDataSessionRegSingleSesion,
		TypeDataSessionRegDisconnect,
		TypeDataMessage,
		TypeDataSend,
		TypeDataSendBytes,
	)
	for _, t := range arr {
		pubsub.Sub(t, fn)
	}
}

func handlerPubSubEvent(data PubSubData, sessionRegistry SessionRegistry, tracker Tracker, logger *zap.Logger) {
	var cmdMsg CmdMessage
	err := json.Unmarshal(data.Data, &cmdMsg)
	if err != nil {
		logger.
			With(zap.Error(err)).
			Error("Parse payload failed")
		return
	}
	// if cmdMsg.SessionId.String() == "" {
	// 	return "", errors.New("Missing session id"), codes.InvalidArgument
	// }
	logger.
		With(zap.String("session", data.SessionId)).
		With(zap.Int("type cmd", int(data.TypeData))).
		Debug("Exec cmd remote sesion")
	sessionId, _ := uuid.FromString(data.SessionId)
	switch data.TypeData {
	case TypeDataRemoveSession:
		sessionRegistry.Remove(sessionId)
	case TypeDataSessionSetUserName:
		// todo
	case TypeDataSessionRegSingleSesion:
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		userID, _ := uuid.FromString(string(cmdMsg.Payload))
		sessionRegistry.SingleSession(ctx,
			tracker, userID, sessionId)
	case TypeDataSessionRegDisconnect:
		sessionDisStruct := SessionDisconnectPayload{}
		err = json.Unmarshal(cmdMsg.Payload, &sessionDisStruct)
		if err != nil {
			logger.
				With(zap.Error(err)).
				Error("Parse payload CmdSessionRegDisconnect")
		}
		listSessionUuid := make([]uuid.UUID, 0, len(sessionDisStruct.SessionId))
		for _, str := range sessionDisStruct.SessionId {
			sessionUUID, _ := uuid.FromString(str)
			listSessionUuid = append(listSessionUuid, sessionUUID)
		}
		listReason := make([]runtime.PresenceReason, 0,
			len(sessionDisStruct.Reasons))
		for _, r := range sessionDisStruct.Reasons {
			listReason = append(listReason, runtime.PresenceReason(r))
		}
		for _, uuid := range listSessionUuid {
			{
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				sessionRegistry.Disconnect(ctx, uuid, listReason...)
			}
		}
	case TypeDataSend:
		session := sessionRegistry.Get(sessionId)
		if session == nil {
			logger.
				With(zap.String("session id", data.SessionId)).
				Error("Session is nil")
		}
		envelope := &rtapi.Envelope{}
		err = json.Unmarshal(cmdMsg.Payload, envelope)
		if err != nil {
			logger.
				With(zap.Error(err)).
				With(zap.ByteString("payload", cmdMsg.Payload)).
				Error("Parse payload rtapi.Envelope failed")
		}
		err = session.Send(envelope, cmdMsg.Reliable)
		if err != nil {
			logger.
				With(zap.String("session id", data.SessionId)).
				With(zap.ByteString("payload", cmdMsg.Payload)).
				With(zap.Error(err)).
				Error("Session Send failed")
		}
	case TypeDataSendBytes:
		session := sessionRegistry.Get(sessionId)
		if session == nil {
			logger.
				With(zap.String("session id", data.SessionId)).
				Error("Session is nil")
		}
		err = session.SendBytes(cmdMsg.Payload, cmdMsg.Reliable)
		if err != nil {
			logger.
				With(zap.String("session id", data.SessionId)).
				With(zap.ByteString("payload", cmdMsg.Payload)).
				With(zap.Error(err)).
				Error("Session Send failed")
		}
	}
}
