package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gorilla/websocket"
	ev "github.com/nguyenzung/nodego/eventloop"
)

type IMessage interface {
	process()
}

type BaseMessage struct {
	MessageType int         `json:"type"`
	Object      interface{} `json:"content"`
}

var (
	LOGIN_REQUEST_TYPE  = 0
	LOGIN_RESPONSE_TYPE = 1

	CALL_REQUEST_TYPE  = 8
	CALL_RESPONSE_TYPE = 9

	OFFER_TYPE  = 16
	ANSWER_TYPE = 17

	ICE_CANDIDATE_TYPE = 24
)

type LoginRequestMessage struct {
}

func (message *LoginRequestMessage) process() {
	fmt.Println("LoginRequestMessage process")
}

type LoginResponseMessage struct {
}

func (message *LoginResponseMessage) process() {
	fmt.Println("LoginResponseMessage process")
}

type CallRequestMessage struct {
}

type CallResponseMessage struct {
}

type OfferMessage struct {
}

type AnswerMessage struct {
}

type ICECandidateMessage struct {
}

type SignalingService struct {
	app *ev.App
}

func onMessage(me *ev.MessageEvent, s *ev.Session) {
	fmt.Println(" onMessage ", me.MessageType, string(me.Data))
	var message BaseMessage
	err := json.Unmarshal(me.Data, &message)
	if err != nil {
		s.WriteText([]byte("Error"))
	} else {
		switch message.MessageType {
		case LOGIN_REQUEST_TYPE:
			break

		case LOGIN_RESPONSE_TYPE:
			break

		case CALL_REQUEST_TYPE:
			break

		case CALL_RESPONSE_TYPE:
			break

		case OFFER_TYPE:
			break

		case ANSWER_TYPE:
			break

		case ICE_CANDIDATE_TYPE:
			break
		}
		s.WriteText([]byte(strconv.Itoa(message.MessageType)))
		str := fmt.Sprintf("%v", message.Object)
		s.WriteText([]byte(str))
	}
}

func onClose(ce *ev.CloseEvent, s *ev.Session) error {
	s.WriteClose(websocket.CloseMessage, "Terminate connection")
	return nil
}

func (service *SignalingService) initController() {
	service.app.MakeWSHandler("/", onMessage, onClose)
}

func (service *SignalingService) Exec() {
	service.app = ev.NewApp()
	service.initController()
	service.app.Exec()
}

func MakeService() *SignalingService {
	return &SignalingService{}
}
