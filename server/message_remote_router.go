package server

import (
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
