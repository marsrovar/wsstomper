package main

import (
	"log"
	"net/url"
	"time"

	stomp "github.com/drawdy/stomp-ws-go"
	"github.com/gorilla/websocket"
)

func main() {
	go topicTest()
	topicUser1()
}

func topicTest() {
	u := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8080",
		Path:   "/",
	}

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("couldn't connect to %v: %v\n", u.String(), err)
	}

	log.Printf("response status: %v\n", resp.Status)
	log.Println("Websocket connection succeeded.")

	h := stomp.Headers{
		stomp.HK_ACCEPT_VERSION, stomp.SPL_12,
		stomp.HK_HEART_BEAT, "3000,3000",
		stomp.HK_HOST, u.Host,
	}
	sc, err := stomp.ConnectOverWS(conn, h)
	if err != nil {
		log.Fatalf("couldn't create stomp connection: %v\n", err)
	}

	scUUID := stomp.Uuid()

	mdCh, err := sc.Subscribe(stomp.Headers{
		stomp.HK_DESTINATION, "/topic/test",
		stomp.HK_ID, scUUID,
	})
	if err != nil {
		log.Fatalf("failed to subscribe greeting message: %v\n", err)
	}

	for {
		select {
		case msg := <-mdCh:
			log.Println("test", time.Now(), msg.Message.BodyString())
			log.Println("------")
		}
	}
}

func topicUser1() {
	u := url.URL{
		Scheme: "ws",
		Host:   "127.0.0.1:8080",
		Path:   "/",
	}
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("couldn't connect to %v: %v\n", u.String(), err)
	}

	log.Printf("response status: %v\n", resp.Status)
	log.Println("Websocket connection succeeded.")

	h := stomp.Headers{
		stomp.HK_ACCEPT_VERSION, stomp.SPL_12,
		stomp.HK_HEART_BEAT, "3000,3000",
		stomp.HK_HOST, u.Host,
	}
	sc, err := stomp.ConnectOverWS(conn, h)
	if err != nil {
		log.Fatalf("couldn't create stomp connection: %v\n", err)
	}

	scUUID := stomp.Uuid()

	mdCh, err := sc.Subscribe(stomp.Headers{
		stomp.HK_DESTINATION, "/topic/test/user1",
		stomp.HK_ID, scUUID,
	})
	if err != nil {
		log.Fatalf("failed to subscribe greeting message: %v\n", err)
	}

	for {
		select {
		case msg := <-mdCh:
			log.Println("user1", time.Now(), msg.Message.BodyString())
			log.Println("+++++")
		}
	}
}
