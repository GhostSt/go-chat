package main

import (
	"fmt"
	"log"
)

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan *Message
	register   chan *Client
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan *Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) Broadcast(message *Message) {
	h.broadcast <- message
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <- h.unregister:
			if _, ok := h.clients[client]; ok {
				log.Print("Closing client if unregistered.")
				delete(h.clients, client)
				close(client.send)
			}
		case message := <- h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
					log.Print(fmt.Sprintf("Sending message to client: %s", message.Body))
				default:
					log.Print("Closing client while broadcasting if unable send message to client.")
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}