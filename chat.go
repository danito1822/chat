// main.go
package main

import (
	"flag"
	"log"
	"net/http"
	"text/template"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", ":8090", "http service address")

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
} // use default options

type Message struct {
	Username string `json:"username"`
	Text     string `json:"text"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func main() {
	flag.Parse()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws", wsHandler)

	go handleMessages()

	log.Println("Servidor iniciado en", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("index.html")
	t.Execute(w, nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clients[ws] = true

	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
