package wsstomper

import (
	"bytes"
	"container/list"
	"regexp"
	"strconv"

	"github.com/go-stomp/stomp/v3/frame"
	"github.com/olahol/melody"
	cmap "github.com/orcaman/concurrent-map/v2"
)

type (
	ConnectCheckHeaderHandle func(header *frame.Header) bool
	NotifyHandler            func(session *melody.Session, contentType string, data []byte)
)

type StompServer struct {
	subscriptions     cmap.ConcurrentMap[string, *list.List]
	subscriptionIdBox cmap.ConcurrentMap[string, *sessionLinkElement]

	connectCheckHandle ConnectCheckHeaderHandle
}

type sessionData struct {
	S              *melody.Session
	SubscriptionId string
}

type sessionLinkElement struct {
	Destination    string
	S              *melody.Session
	SubscriptionId string
	Element        *list.Element
}

func NewStomp() *StompServer {
	return &StompServer{
		subscriptions:     cmap.New[*list.List](),
		subscriptionIdBox: cmap.New[*sessionLinkElement](),
	}
}

func (stompS *StompServer) AddConnectCheckHandle(handler ConnectCheckHeaderHandle) {
	stompS.connectCheckHandle = handler
}

func (stompS *StompServer) AddSubscription(destination, subscriptionId string, session *melody.Session) {
	stompS.addSubscriptions(destination, subscriptionId, session)
}

func (stompS *StompServer) addSubscriptions(destination, subscriptionId string, session *melody.Session) {
	link, ok := stompS.subscriptions.Get(destination)
	if !ok {
		link = list.New()

		stompS.subscriptions.Set(destination, link)
	}

	e := link.PushBack(&sessionData{
		S:              session,
		SubscriptionId: subscriptionId,
	})

	stompS.subscriptionIdBox.Set(subscriptionId, &sessionLinkElement{
		Destination:    destination,
		S:              session,
		SubscriptionId: subscriptionId,
		Element:        e,
	})
}

func (stompS *StompServer) UnSubscription(subscriptionId string) {
	linkElement, ok := stompS.subscriptionIdBox.Get(subscriptionId)
	if !ok {
		return
	}

	link, ok := stompS.subscriptions.Get(linkElement.Destination)
	if !ok {
		stompS.subscriptionIdBox.Remove(subscriptionId)
		return
	}

	if link.Len() == 0 {
		return
	}

	link.Remove(linkElement.Element)
}

func (stompS *StompServer) SendMessageWithCheck(destination, contentType string, body []byte, check func(session *melody.Session) bool) error {
	link, ok := stompS.subscriptions.Get(destination)
	if !ok {
		return nil
	}

	for node := link.Front(); node != nil; node = node.Next() {
		s := node.Value.(*sessionData)

		if s.S.IsClosed() {
			link.Remove(node)
			continue
		}

		if check != nil && !check(s.S) {
			continue
		}

		var b bytes.Buffer
		if err := frame.NewWriter(&b).Write(&frame.Frame{
			Command: frame.MESSAGE,
			Header: frame.NewHeader(
				frame.Subscription, s.SubscriptionId,
				frame.Destination, destination,
				frame.ContentType, contentType,
				frame.ContentLength, strconv.Itoa(len(body)),
			),
			Body: body,
		}); err != nil {
			return err
		}

		s.S.Write(b.Bytes())
	}

	if link.Len() == 0 {
		stompS.subscriptions.Remove(destination)
	}

	return nil
}

func (stompS *StompServer) SendMessage(destination, contentType string, body []byte) {
	stompS.SendMessageWithCheck(destination, contentType, body, nil)
}

func (stompS *StompServer) SendMessageByRegex(regex, contentType string, body []byte) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return
	}

	for _, destination := range stompS.subscriptions.Keys() {
		if !r.MatchString(destination) {
			continue
		}

		stompS.SendMessage(destination, contentType, body)
	}
}

func (stompS *StompServer) NotifyAllSession(contentType string, body []byte) {
	for _, destination := range stompS.subscriptions.Keys() {
		stompS.SendMessage(destination, contentType, body)
	}
}

func (stompS *StompServer) SendError(subscriptionId string, body []byte) error {
	linkElement, ok := stompS.subscriptionIdBox.Get(subscriptionId)
	if !ok {
		return nil
	}

	var b bytes.Buffer
	if err := frame.NewWriter(&b).Write(&frame.Frame{
		Command: frame.ERROR,
		Body:    body,
	}); err != nil {
		return err
	}

	return linkElement.S.Write(b.Bytes())
}
