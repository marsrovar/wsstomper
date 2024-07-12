package wsstomper

import (
	"bytes"
	"errors"
	"strings"

	"github.com/go-stomp/stomp/v3"
	"github.com/go-stomp/stomp/v3/frame"
	"github.com/olahol/melody"
)

func (stompS *StompServer) Command(session *melody.Session, msg []byte) error {
	f, err := frame.NewReader(strings.NewReader(string(msg))).Read()
	if err != nil {
		return err
	}

	switch f.Command {
	case frame.CONNECT:
		if stompS.connectCheckHandle != nil && !stompS.connectCheckHandle(f.Header) {
			return errors.New("connect fail")
		}

		return stompS.connectCommand(session, f)
	case frame.SEND:
		return stompS.sendCommand(session, msg, f)
	case frame.SUBSCRIBE:
		return stompS.subscribeCommand(session, f)
	case frame.UNSUBSCRIBE:
		return stompS.unsubscribeCommand(session, f)
	default:
		return errors.New("not sup support: " + f.Command)
	}
}

func (stompS *StompServer) connectCommand(session *melody.Session, f *frame.Frame) error {
	v := f.Header.Get(frame.AcceptVersion)
	version := strings.Split(v, ",")

	for _, v := range version {
		if err := stomp.Version(v).CheckSupported(); err != nil {
			session.Close()
			return err
		}
	}

	var b bytes.Buffer
	if err := frame.NewWriter(&b).Write(&frame.Frame{
		Command: frame.CONNECTED,
		Header: frame.NewHeader(
			frame.Version, v,
		),
	}); err != nil {
		return err
	}

	return session.Write(b.Bytes())
}

func (stompS *StompServer) subscribeCommand(session *melody.Session, f *frame.Frame) error {
	destination := f.Header.Get(frame.Destination)
	if destination == "" {
		return errors.New("must need destination")
	}

	id := f.Header.Get(frame.Id)
	if id == "" {
		return errors.New("must need id")
	}

	if _, exists := session.Get(id); exists {
		return errors.New("duplicate subscription-id")
	}

	session.Set(SubscriptionIdKey, id)

	stompS.AddSubscription(destination, id, session)

	return nil
}

func (stompS *StompServer) unsubscribeCommand(session *melody.Session, f *frame.Frame) error {
	subscriptionId := f.Header.Get(frame.Id)
	if subscriptionId == "" {
		return errors.New("must need id")
	}

	stompS.UnSubscription(subscriptionId)

	return nil
}

func (stompS *StompServer) sendCommand(session *melody.Session, msg []byte, f *frame.Frame) error {
	destination := f.Header.Get(frame.Destination)
	if destination == "" {
		return errors.New("must need destination")
	}

	contentType := f.Header.Get(frame.ContentType)
	if contentType == "" {
		contentType = ContentTypeText
	}

	stompS.SendMessage(destination, contentType, msg)

	return nil
}
