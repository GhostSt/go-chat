package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	type_success = "0x00001"
	type_fail    = "0x00002"
)

type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan *Message
}

type Message struct {
	Code string `json:"code"`
	Type string `json:"type"`
	Body string `json:"body"`
}

// Reads messages from connection and pumps them to hub
func (c *Client) Read() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()

		if err != nil {
			log.Print(fmt.Printf("notice: %s", err))
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, rawMessage, err := c.conn.ReadMessage()

		if err != nil {
			log.Print(fmt.Sprintf("notice: couldn't read message: %s", err))
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}

			// TODO: add handling another close status codes
			return
		}

		message := &Message{}

		err = json.Unmarshal(rawMessage, message)

		if err != nil {
			log.Print(fmt.Sprintf("Unable to parse message: %s", rawMessage))
		}

		c.hub.Broadcast(message)
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		err := c.conn.Close()

		if err != nil {
			log.Print(fmt.Printf("notice: %s", err))
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				log.Print("Closing connection")

				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)

			if err != nil {
				log.Printf("notice: error while getting writer - %v", err)
				return
			}

			rawMessage, _ := json.Marshal(message)

			c.write(w, rawMessage)

			n := len(c.send)

			for i := 0; i < n; i++ {
				c.write(w, newline)

				rawMessage, _ := json.Marshal(<-c.send)
				c.write(w, rawMessage)
			}

			if err := w.Close(); err != nil {
				log.Printf("notice: error while closing writer - %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("notice: %v", err)

				return
			}
		}
	}
}

func (c *Client) write(w io.Writer, message []byte) {
	_, err := w.Write(message)

	if err != nil {
		log.Print(fmt.Sprintf("notice: unable to write message to connection writer: %s", err))
	}
}
