package server

import (
	"github.com/heroiclabs/nakama-common/rtapi"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
)

type RemoteMessageRouter struct {
	LocalMessageRouter
}

func NewRemoteMessageRouter(sessionRegistry SessionRegistry, tracker Tracker, protojsonMarshaler *protojson.MarshalOptions) MessageRouter {
	r := &RemoteMessageRouter{
		LocalMessageRouter: LocalMessageRouter{
			protojsonMarshaler: protojsonMarshaler,
			sessionRegistry:    sessionRegistry,
			tracker:            tracker,
		},
	}

	return r
}
func (r *RemoteMessageRouter) SendToPresenceIDs(logger *zap.Logger, presenceIDs []*PresenceID, envelope *rtapi.Envelope, reliable bool) {
	if len(presenceIDs) == 0 {
		return
	}
	localPresence := make([]*PresenceID, 0)
	remotePresence := make([]*PresenceID, 0)

	for _, presenceID := range presenceIDs {
		session := r.sessionRegistry.Get(presenceID.SessionID)
		v := presenceID
		if session == nil {
			// logger.Debug("No session to route to", zap.String("sid", presenceID.SessionID.String()))
			remotePresence = append(remotePresence, v)
			continue
		}
		localPresence = append(localPresence, v)
	}
	if len(localPresence) > 0 {
		r.LocalMessageRouter.SendToPresenceIDs(logger, localPresence, envelope, reliable)
	}
	if len(remotePresence) > 0 {
		r.SendToRemotePresenceIDs(logger, remotePresence, envelope, reliable)
	}
}

func (r *RemoteMessageRouter) SendToRemotePresenceIDs(logger *zap.Logger, presenceIDs []*PresenceID, envelope *rtapi.Envelope, reliable bool) {
}
