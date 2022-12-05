package server

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
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
	Node      string
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

func NewPubSubDataFromProtoMsg(node string, typeData TypeData, sessionId string, data proto.Message) PubSubData {
	p := PubSubData{
		Node:      node,
		TypeData:  typeData,
		SessionId: sessionId,
	}
	p.Data, _ = jsonpbMarshaler.Marshal(data)
	return p
}

type PubSubEvent interface {
	Pub(PubSubData) error
	Sub(TypeData, FnSubEvent)
}

type PubSubHandler struct {
	redisClient   *redis.Client
	mapFnSubEvent map[TypeData][]FnSubEvent
}

func NewPubSubHandler(redisClient *redis.Client, node string) PubSubEvent {
	p := &PubSubHandler{
		redisClient: redisClient,
	}
	p.mapFnSubEvent = make(map[TypeData][]FnSubEvent, 0)

	if p.redisClient == nil {
		return p
	}
	go func() {
		ctx := context.Background()
		subscriber := p.redisClient.Subscribe(ctx, node)
		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				return
			}
			var data PubSubData
			err = json.Unmarshal([]byte(msg.Payload), &data)
			if err != nil {
				continue
			}
			listFn, exist := p.mapFnSubEvent[data.TypeData]
			if !exist {
				continue
			}
			for _, fn := range listFn {
				fn(data)
			}
		}
	}()
	return p
}

func (p *PubSubHandler) Pub(pubData PubSubData) error {
	if p.redisClient == nil {
		return errors.New("redis client is nil")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := p.redisClient.Publish(ctx, pubData.Node, pubData).Result()
	if err != nil {
		zap.L().With(zap.String("node", pubData.Node)).
			With(zap.Error(err)).Error("publish message failed")
		return err
	}
	return nil
}

func (p *PubSubHandler) Sub(typeData TypeData, fnCallBack FnSubEvent) {
	v, exist := p.mapFnSubEvent[typeData]
	if !exist {
		v = make([]FnSubEvent, 0)
	}
	v = append(v, fnCallBack)
	p.mapFnSubEvent[typeData] = v
}
