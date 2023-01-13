package main

import (
	"encoding/json"
	"fmt"

	"github.com/mitchellh/mapstructure"
	ev "github.com/nguyenzung/nodego/eventloop"
)

type BaseMessage struct {
	MessageType int         `json:"type"`
	Object      interface{} `json:"content"`
}

var (
	LOGIN_REQUEST_TYPE  = 0
	LOGIN_RESPONSE_TYPE = 1

	CALL_REQUEST_TYPE        = 32 // Request a call to an id
	CALL_RESPONSE_TYPE       = 33 // Reponse to the call's request
	CALL_REQUEST_STATUS_TYPE = 34 // Status of a request after processing
	CALL_REQUEST_CANCEL_TYPE = 35 // Caller cancel call request

	ICE_INFO_TYPE = 64

	OFFER_TYPE  = 96
	ANSWER_TYPE = 97

	ICE_CANDIDATE_TYPE = 120
)

// type iMessage interface {
// 	LoginRequestMessage | LoginResponseMessage | CallRequestMessage | CallRequestStatusMessage | CallRequestCancelMessage | CallResponseMessage | CallSetupMessage
// }

type LoginRequestMessage struct {
}

type LoginResponseMessage struct {
	Id uint64 `json:"id"`
}

func MakeLoginResponseMessage(id uint64) *LoginResponseMessage {
	return &LoginResponseMessage{Id: id}
}

type CallRequestMessage struct {
	CalleeId uint64 `json:"calleeId"`
}

type CallRequestStatusMessage struct {
	Status   bool   `json:"status"`
	CallerId uint64 `json:"callerId"`
	CalleeId uint64 `json:"calleeId"`
}

func MakeCallRequestStatusMessage(status bool, caller uint64, callee uint64) *CallRequestStatusMessage {
	return &CallRequestStatusMessage{status, caller, callee}
}

type CallRequestCancelMessage struct {
}

type CallRequestCancelStatusMessage struct {
}

type CallResponseMessage struct {
	CallerId uint64 `json:"callerId"`
	Status   bool   `json:"status"`
}

type CallSetupMessage struct {
	Ices []string `json:"ices"`
}

type CallSetupAckMessage struct {
}

type CallCancelMessage struct {
}

type SetupCallMessage struct {
	Config string `json:"config"`
}

type CallRoom struct {
	CallerAck uint64 `json:"callerAck"`
	CalleeAck uint64 `json:"calleeAck"`
	CallerId  uint64 `json:"callerId"`
	CalleeId  uint64 `json:"calleeId"`
}

type SignalingService struct {
	connections            map[uint64]*ev.Session
	connectionIdBySessions map[*ev.Session]uint64
	sendingRequests        map[uint64]uint64 // store request from caller to callee
	receivingRequests      map[uint64]uint64 // store request to callee from caller
	callingPairs           map[uint64]*CallRoom
	maxID                  uint64
}

func (service *SignalingService) sendMessageToClient(socketId uint64, message any) {
	connection, ok := service.connections[socketId]
	if !ok {
		return
	}
	b, e := json.Marshal(message)
	if e == nil {
		connection.WriteText(b)
	}
}

func (service *SignalingService) onOpen(s *ev.Session) {
	service.connections[service.maxID] = s
	service.connectionIdBySessions[s] = service.maxID
	message := MakeLoginResponseMessage(service.maxID)
	service.sendMessageToClient(service.maxID, message)
	service.maxID++
}

func (service *SignalingService) onClose(ce *ev.CloseEvent, s *ev.Session) error {
	peerId := service.connectionIdBySessions[s]
	delete(service.connections, peerId)
	delete(service.connectionIdBySessions, s)
	delete(service.sendingRequests, peerId)
	delete(service.receivingRequests, peerId)
	delete(service.callingPairs, peerId)
	return nil
}

func (service *SignalingService) processCallRequest(message *CallRequestMessage, s *ev.Session) {
	fmt.Println("Request call to", message.CalleeId)
	callerId := service.connectionIdBySessions[s]
	_, isInARequest := service.sendingRequests[callerId]
	_, isInACall := service.callingPairs[callerId]
	_, isCalleeAvailable := service.connections[message.CalleeId]
	_, isCalleeOnRequest := service.receivingRequests[message.CalleeId]
	_, isCalleeOnCall := service.callingPairs[message.CalleeId]

	fmt.Println(" Call Request DB", isInARequest, isInACall, isCalleeAvailable, isCalleeOnRequest)

	// An user can make a request if he/she is not in a request making or a call and callee is available and not in a call
	if !isInARequest && !isInACall && isCalleeAvailable && !isCalleeOnRequest && !isCalleeOnCall && callerId != message.CalleeId {
		service.sendingRequests[callerId] = message.CalleeId
		service.receivingRequests[message.CalleeId] = callerId
		message := MakeCallRequestStatusMessage(true, callerId, message.CalleeId)
		service.sendMessageToClient(callerId, message)
		service.sendMessageToClient(message.CalleeId, message)
	} else {
		message := MakeCallRequestStatusMessage(false, callerId, message.CalleeId)
		service.sendMessageToClient(callerId, message)
		fmt.Println("callerId is in a request call or in a call")
	}
}

func (service *SignalingService) processCallResponse(message *CallResponseMessage, s *ev.Session) {
	fmt.Println("Request call to", message.CallerId, message.Status)
	callerId := message.CallerId
	calleeId := service.connectionIdBySessions[s]
	status := message.Status
	fmt.Println("CallResponse ", calleeId, callerId, status)
}

func (service *SignalingService) onMessage(me *ev.MessageEvent, s *ev.Session) {
	var baseMessage BaseMessage
	err := json.Unmarshal(me.Data, &baseMessage)
	if err != nil {
		s.WriteText([]byte(err.Error()))
	} else {
		switch baseMessage.MessageType {

		case CALL_REQUEST_TYPE:
			var message CallRequestMessage
			mapstructure.Decode(baseMessage.Object, &message)
			service.processCallRequest(&message, s)

		case CALL_RESPONSE_TYPE:
			var message CallResponseMessage
			mapstructure.Decode(baseMessage.Object, &message)
			service.processCallResponse(&message, s)

		case OFFER_TYPE:
			break

		case ANSWER_TYPE:
			break

		case ICE_CANDIDATE_TYPE:
			break
		}
	}
}

func (service *SignalingService) initController() {
	ev.MakeWSHandler("/", service.onOpen, service.onMessage, service.onClose)
}

func (service *SignalingService) Exec() {
	ev.InitApp()
	service.connections = make(map[uint64]*ev.Session)
	service.connectionIdBySessions = make(map[*ev.Session]uint64)
	service.sendingRequests = make(map[uint64]uint64)
	service.receivingRequests = make(map[uint64]uint64)
	service.callingPairs = make(map[uint64]*CallRoom)
	service.maxID = 0
	service.initController()
	ev.ExecApp()
}

func MakeService() *SignalingService {
	return &SignalingService{}
}
