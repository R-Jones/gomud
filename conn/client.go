// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package conn

import (
	"bytes"
	"log"
	"net/http"
	"time"
        "fmt"
        "../model"
        "../cmd"
	"github.com/gorilla/websocket"
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
        CheckOrigin: func(r *http.Request) bool {
            return true
        },
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
        player *model.Player
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
                retval := processMessage(c,message)
                fmt.Println("readpump",retval)
		c.hub.broadcast <- []byte(retval)
	}
}

func processMessage(c *Client, message []byte) (retval string){
                        split := bytes.Split(message, []byte("`"))
                        //fmt.Println("%q\n",split)
                        allparams := [][]byte{}//[]byte is a string, thus [][]byte is an array of strings(i.e. our parameters)
                        for i, param := range split {
                            if i % 2 == 0 {

                                if trimmed := bytes.TrimSpace(param); len(trimmed) > 0 {
                                    splitted := bytes.Split(trimmed, []byte(" "))
                                    allparams = append(allparams, splitted...)
                                }
                            } else {
                                allparams = append(allparams, param)
                            }
                        }
                        //Can refactor the below into a new map[[]byte]func
                        if(bytes.Equal(allparams[0],[]byte("dig"))) {
                            cmd.Dig(c.player,allparams...)
                        }
                        if(bytes.Equal(allparams[0],[]byte("look"))) {
                            retval = cmd.Look(c.player,allparams...)
                        }
                        //fmt.Println(string(message))
                        //fmt.Printf("%q\n",allparams)
                        fmt.Println(string(allparams[0]))
                        return retval
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
                        fmt.Println("writePump",message)
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
                        fmt.Sprintf("%T", message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
        
        player := model.CreatePlayer(&model.Player_{})
        fmt.Println(player)
	client := &Client{hub: hub, conn: conn, player:player, send: make(chan []byte, 256)}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}