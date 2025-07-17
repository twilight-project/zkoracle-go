package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

var upgrader = websocket.Upgrader{}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.conn.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, _ := upgrader.Upgrade(w, r, nil)
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
}

func pubsubServer() {
	WsHub = &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}

	go WsHub.run()

	http.HandleFunc("/latestblock", func(w http.ResponseWriter, r *http.Request) {
		// Parse the query parameters
		r.ParseForm()
		blockHeightStr := r.Form.Get("blockHeight")

		var blockHeight int64 = 0 // default value
		if blockHeightStr != "" {
			var err error
			blockHeight, err = strconv.ParseInt(blockHeightStr, 10, 64)
			if err != nil {
				http.Error(w, "Invalid block height parameter", http.StatusBadRequest)
				return
			}
		}

		// Pass the int64 parameter to the serveWs function
		go serveWs(WsHub, w, r)
		nyksSubscriber(uint64(blockHeight))
	})

	err := http.ListenAndServe(":7001", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
