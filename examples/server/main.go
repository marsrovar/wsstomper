package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/marsrovar/wsstomper"
	"github.com/olahol/melody"
)

var (
	IWs    = melody.New()
	IStomp = wsstomper.NewStomp()
)

func main() {
	go Worker()

	IWs.HandleConnect(onConnectSTOMPEvent)
	IWs.HandleDisconnect(onDisconnectSTOMPEvent)
	IWs.HandleClose(onCloseSTOMPEvent)
	IWs.HandleMessage(onMessageSTOMPEvent)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		IConnectSTOMP(w, r)
	})

	http.ListenAndServe(":8080", nil)
}

func IConnectSTOMP(w http.ResponseWriter, r *http.Request) error {
	return IWs.HandleRequestWithKeys(w, r, map[string]interface{}{})
}

func onConnectSTOMPEvent(s *melody.Session) {
	log.Println("connect:", s.RemoteAddr().String())
}

func onDisconnectSTOMPEvent(s *melody.Session) {
	log.Println("disconnect:", s.RemoteAddr().String())

	id, ok := s.Get(wsstomper.SubscriptionIdKey)
	if !ok {
		return
	}

	if _, ok := id.(string); !ok {
		return
	}

	IStomp.UnSubscription(id.(string))
}

func onCloseSTOMPEvent(s *melody.Session, code int, msg string) error {
	log.Println("close:", s.RemoteAddr().String())
	return nil
}

func onMessageSTOMPEvent(s *melody.Session, data []byte) {
	if err := IStomp.Command(s, data); err != nil {
		log.Println("message error:", err.Error())
	}
}

func Worker() {
	type Test struct {
		Number int `json:"number"`
	}

	count := 0
	for {
		b, _ := json.Marshal(Test{count})
		IStomp.SendMessage("/topic/test/user1", wsstomper.ContentTypeJson, b)
		IStomp.SendMessageByRegex("^/topic/test.*$", wsstomper.ContentTypeJson, b)
		count++

		time.Sleep(time.Second)
	}
}
