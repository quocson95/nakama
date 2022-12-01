package server

import (
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

type TypeData int

const (
	TypeDataPipeMatchCreate   TypeData = 1
	TypeDataPipeMatchDataSend TypeData = 2
	TypeDataPipeMatchJoin     TypeData = 3
	TypeDataPipeMatchLeave    TypeData = 4
)

type FnSubEvent func(PubSubData)
type PubSubData struct {
	TypeData  TypeData
	SessionId string
	Data      []byte
}

var jsonpbUnmarshaler = &protojson.UnmarshalOptions{
	DiscardUnknown: false,
}
var jsonpbMarshaler = &protojson.MarshalOptions{
	UseEnumNumbers:  true,
	EmitUnpopulated: false,
	Indent:          "",
	UseProtoNames:   true,
}

func NewPubSubDataFromProtoMsg(typeData TypeData, sessionId string, data proto.Message) PubSubData {
	p := PubSubData{
		TypeData:  typeData,
		SessionId: sessionId,
	}
	p.Data, _ = jsonpbMarshaler.Marshal(data)
	return p
}

type PubSubEvent interface {
	Pub(PubSubData)
	Sub(TypeData, FnSubEvent)
}
