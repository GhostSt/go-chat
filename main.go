package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func serveFrontPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Page not found", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, "public/index.html")
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	log.Print("getting request")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Fatal("Unable to upgrade connection: ", err)
	}

	client := &Client{
		hub: hub,
		conn: conn,
		send: make(chan *Message, 256),
	}

	hub.Register(client)

	go client.Write()
	go client.Read()

	time.Sleep(50 * time.Millisecond)
	activeClients := len(hub.clients)

	log.Print(fmt.Sprintf("Active clients: %d", activeClients))
	message := &Message{
		Code: type_success,
		Type: "online",
		Body: strconv.Itoa(activeClients),
	}
	hub.Broadcast(message)
}

func main() {
	hub := newHub()
	go hub.Run()

	log.Print("starting http server")
	http.HandleFunc("/", serveFrontPage)
	http.HandleFunc("/ws", func (w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})

	err := http.ListenAndServe(":8001", nil)

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
