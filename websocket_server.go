package main

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Client struct {
	ID       int
	Conn     *websocket.Conn
	LastRecv time.Time
}

const (
	TIME_OUT = time.Second * 60
)

var (
	ClientHub  []Client
	timeToVoid time.Duration = time.Second * 60
)

type WebsocketEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

func DeleteClient(ID int, err error) {
	for i := range ClientHub {
		if ClientHub[i].ID == ID {
			clientID := ClientHub[i].ID
			ClientHub[i].Conn.Close()
			ClientHub = append(ClientHub[:i], ClientHub[i+1:]...)
			fmt.Printf("Disconnected client %d.\n CLIENT ERR: %s", clientID, err)
			return
		}
	}
}

func WebsocketHandler() {
	for {
		for _, y := range ClientHub {
			msgType, msg, err := y.Conn.ReadMessage()
			if err != nil {
				DeleteClient(y.ID, err)
				continue
			}

			resp := WebsocketEvent{}
			err = JSONDecode(msg, &resp)
			log.Println("AFTER DECODE RESP>: ", resp)
			if err != nil {
				log.Printf("Recv'd a plain text event: %s\n", (string(msg)))
				if string(msg) == "ping" {
					y.LastRecv = time.Now()
					log.Println("Ping Recieved")
					err = y.Conn.WriteMessage(msgType, []byte("pong"))
					if err != nil {
						DeleteClient(y.ID, err)
						continue
					}
				}
			} else {
				log.Printf("Recv'd a %s WS event", resp.Event)
				switch resp.Event {
				case "tickerrequest":
					log.Println("resp.Data in TICKERREQUEST: ", resp.Data)
					ticker, err := JSONEncode("TestOUTGOING")
					if err != nil {
						log.Println(err)
						continue
					}
					y.Conn.WriteMessage(msgType, ticker)
				}
			}
			timeResultant := y.LastRecv.Sub(time.Now())
			if timeToVoid < timeResultant {
				err := errors.New("Connection Lost")
				DeleteClient(y.ID, err)
			}
		}
	}
}

func WebsocketConn(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		WriteBufferSize: 1024,
		ReadBufferSize:  1024,
	}

	newClient := Client{}
	newClient.ID = len(ClientHub)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
	}

	newClient.Conn = conn
	ClientHub = append(ClientHub, newClient)
	log.Printf("Recv'd new WS client: ID %d (total clients (%d)\n", newClient.ID, len(ClientHub))
}

func ServeTestIndex(w http.ResponseWriter, r *http.Request) {
	buffer, err := ioutil.ReadFile("wankerfolder/index.html")
	if err != nil {
		fmt.Println(err)
		return
	}
	w.Write(buffer)
}
