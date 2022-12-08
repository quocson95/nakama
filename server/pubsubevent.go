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
	TypeDataPipeMatchList TypeData = iota
	TypeDataPipeMatchCreate
	TypeDataPipeMatchDataSend
	TypeDataPipeMatchJoin
	TypeDataPipeMatchLeave

	TypeDataRemoveSession
	TypeDataSessionSetUserName
	TypeDataSessionRegSingleSesion
	TypeDataSessionRegDisconnect
	TypeDataMessage
	TypeDataSend
	TypeDataSendBytes
)

type FnSubEvent func(PubSubData)
type PubSubData struct {
	Node      string
	TypeData  TypeData
	SessionId string
	Data      []byte
}

func (psd PubSubData) MarshalBinary() ([]byte, error) {
	return json.Marshal(psd)
}

type CmdMessage struct {
	Reliable bool   `json:"reliable"`
	Payload  []byte `json:"payload"`
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
	Stop()
	Pub(PubSubData) error
	PubInf(PubSubData, interface{}) error
	Sub(TypeData, FnSubEvent)
}

type PubSubHandler struct {
	redisClient   *redis.Client
	mapFnSubEvent map[TypeData][]FnSubEvent
	ctx           context.Context
	logger        *zap.Logger
}

func NewPubSubHandler(redisClient *redis.Client, logger *zap.Logger, node string) PubSubEvent {
	p := &PubSubHandler{
		redisClient: redisClient,
		ctx:         context.Background(),
		logger:      logger,
	}
	p.mapFnSubEvent = make(map[TypeData][]FnSubEvent, 0)

	if p.redisClient == nil {
		return p
	}
	go func() {
		isStop := false
		for {
			if isStop {
				return
			}
			subscriber := p.redisClient.Subscribe(p.ctx, node)
			if _, err := subscriber.Receive(p.ctx); err != nil {
				p.logger.With(zap.Error(err)).
					Error("failed to receive from control PubSub")
				time.Sleep(5 * time.Second)
				continue
			}
			controlCh := subscriber.Channel()
			p.logger.Info("start listening on control PubSub")
			for {
				select {
				case <-p.ctx.Done():
					isStop = true
					break
				case msg := <-controlCh:
					{
						// for msg := range controlCh {
						p.logger.With(zap.String("payload", msg.Payload)).Info("recv messag")
						var data PubSubData
						err := json.Unmarshal([]byte(msg.Payload), &data)
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
				}
			}
		}

	}()
	return p
}

func (p *PubSubHandler) Pub(pubData PubSubData) error {
	if p.redisClient == nil {
		return errors.New("redis client is nil")
	}
	ctx, cancel := context.WithTimeout(p.ctx, 5*time.Second)
	defer cancel()
	_, err := p.redisClient.Publish(ctx, pubData.Node, pubData).Result()
	if err != nil {
		p.logger.With(zap.String("node", pubData.Node)).
			With(zap.Error(err)).Error("publish message failed")
		return err
	}
	p.logger.With(zap.String("node", pubData.Node)).
		With(zap.Error(err)).Info("publish message success")
	return nil
}

func (p *PubSubHandler) PubInf(data PubSubData, body interface{}) error {
	d, _ := json.Marshal(body)
	data.Data = d
	return p.Pub(data)
}

func (p *PubSubHandler) Sub(typeData TypeData, fnCallBack FnSubEvent) {
	v, exist := p.mapFnSubEvent[typeData]
	if !exist {
		v = make([]FnSubEvent, 0)
	}
	v = append(v, fnCallBack)
	p.mapFnSubEvent[typeData] = v
}

func (p *PubSubHandler) Stop() {
	p.ctx.Done()
}
