package websocket

import (
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/gorilla/websocket"
	"io/ioutil"
	"log"
	"time"
)

type InfoSocket struct {
	Token            string
	Empresa          *int64
	Tipo             string
	SoloCambioEstado bool
	Uuid             string
	Id               string
}

type Connection struct {
	// The websocket connection.
	Ws *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// Interface para data misc usada por registro y desregistro

	UserData interface{}

	//
	Wresponse *rest.ResponseWriter
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
L:
	for {
		select {
		case message := <-c.Send:
			err := c.Ws.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				fmt.Printf("Error escribiendo. esto matara al writer")
				break L

			}

		}
	}

	/*for message := range c.Send {
		if err != nil {
			fmt.Printf("Error escribiendo. esto matara al writer")
			break
		}
	}*/
	c.Ws.Close()
}

type Hub struct {
	// Conexiones
	Connections map[*Connection]bool

	// el menaje a repetir en todas los WS. esto podria
	Broadcast chan interface{}

	Ping chan []byte //canal para el ping

	FuncionBr func(*Hub, interface{})

	FuncRegister   func(*Connection)
	FuncUnregister func(*Connection)
	// registro de nuevos sockets
	Register chan *Connection

	// desregistro
	Unregister chan *Connection

	Log *log.Logger
}

func (hb *Hub) Run() {

	if hb.Log == nil {
		hb.Log = log.New(ioutil.Discard, "log-websocket", 0)
	}

	for {
		select {
		case c := <-hb.Register:
			hb.Log.Printf("Regitro %p\n", c)
			hb.Connections[c] = true
			hb.FuncRegister(c)
			hb.Log.Printf("Fin Regitro %p\n", c)

		case c := <-hb.Unregister:
			hb.Log.Printf("DesRegitro %p\n", c)

			hb.FuncUnregister(c)
			if _, ok := hb.Connections[c]; ok {
				delete(hb.Connections, c)
				close(c.Send)
			}
			hb.Log.Printf("Fin DesRegitro %p\n", c)

		case m := <-hb.Broadcast:
			hb.Log.Printf("inicio BR %p\n")

			hb.FuncionBr(hb, m)
			hb.Log.Printf("Fin Br %p\n")

		case pm := <-hb.Ping: //lo que llega al canal ping se envia a todos lados
			hb.Log.Printf("Nuevo mensaje para todos ping\n")
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
