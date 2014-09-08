package websocket

import (
	"fmt"
	"github.com/gorilla/websocket"
	"time"
)

type Connection struct {
	// The websocket connection.
	Ws *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte
}

func (c *Connection) Reader(h *Hub, funcrecv func(*Hub, []byte)) {
	for {
		_, message, err := c.Ws.ReadMessage()
		if err != nil {
			fmt.Printf("Error de lectura del cliente socket cerrado?%s\n", err)
			break
		}
		funcrecv(h, message)

	}
	c.Ws.Close()
}

func (c *Connection) Writer() {
	for message := range c.Send {
		err := c.Ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			fmt.Printf("Error escribiendo. esto matara al writer")
			break
		}
	}
	c.Ws.Close()
}

type Hub struct {
	// Conexiones
	Connections map[*Connection]bool

	// el menaje a repetir en todas los WS. esto podria
	Broadcast chan interface{}

	Ping chan []byte //canal para el ping

	FuncionBr func(*Hub, interface{})

	// registro de nuevos sockets
	Register chan *Connection

	// desregistro
	Unregister chan *Connection
}

func (hb *Hub) Run() {

	for {
		select {
		case c := <-hb.Register:
			hb.Connections[c] = true
		case c := <-hb.Unregister:
			if _, ok := hb.Connections[c]; ok {
				delete(hb.Connections, c)
				close(c.Send)
			}
		case m := <-hb.Broadcast:

			hb.FuncionBr(hb, m)

		case pm := <-hb.Ping: //lo que llega al canal ping se envia a todos lados
			fmt.Printf("Nuevo mensaje para todos ping\n")
			for c := range hb.Connections {
				c.Send <- pm
			}

		}
	}
}
func (hb *Hub) PingB() {
	for {

		time.Sleep(time.Second * 10)

		msg := `{"type":"message","data":"ESTO ES UN PING"}`
		hb.Ping <- []byte(msg)

	}

}
