package main

import (
	"fmt"
	ws "github.com/gorilla/websocket"
	"github.com/ipartner/websocket"
	"net/http"
)

var HUB = websocket.Hub{
	Broadcast:   make(chan interface{}),
	Ping:        make(chan []byte),
	Register:    make(chan *websocket.Connection),
	Unregister:  make(chan *websocket.Connection),
	Connections: make(map[*websocket.Connection]bool),
	FuncionBr:   Mensajes,
}

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     checkOrigin,
}

func Mensajes(h *websocket.Hub, msg interface{}) {

	switch msg.(type) {
	case string:

		fmt.Print("es string")
	case []byte:
		fmt.Printf("Soncadenas de byts")

	}

}

func LeodeCliente(h *websocket.Hub, msg []byte) {
	//aca veras qe hacer con lo que llega del cliente

	fmt.Printf("Leo del cliente!! %s\n", string(msg))

}

func main() {
	go HUB.Run()
	go HUB.PingB()

	http.HandleFunc("/wsocket", handlerWS)
	http.ListenAndServe(":8083", nil)

}

func checkOrigin(r *http.Request) bool {
	return true
}

func handlerWS(w http.ResponseWriter, r *http.Request) {
	fmt.Println("lleog algo")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	c := &websocket.Connection{Ws: conn, Send: make(chan []byte)} //creamos la conexion
	HUB.Register <- c                                             //registamos la conexion
	defer func() { HUB.Unregister <- c }()                        //en caso que la ultima rutina(no GO routina) muera quitamos la conexion de la lista
	go c.Writer()                                                 //se lanza la goroutine a ccargo de la escritura
	c.Reader(&HUB, LeodeCliente)                                  //la goroutine que inicio la conexion se queda leyendo para siempre . si muere se aplica el defer de arriba.

}
