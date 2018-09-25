package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var host = flag.String("host", "localhost", "http service hostess")
var port = flag.String("port", "8000", "http service hostess")
var clients = make(map[*websocket.Conn]string)
var upgrader = websocket.Upgrader{}
var broadcast = make(chan []byte)

func handleMessages() {
	for {
		msg := <-broadcast
		for client := range clients {
			err := client.WriteMessage(1, msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()
	clients[ws] = r.Header.Get("nick")
	msg := []byte(fmt.Sprintf("server: %s is online", clients[ws]))
	broadcast <- msg
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			msg := []byte(fmt.Sprintf("server: %s is offline", clients[ws]))
			broadcast <- msg
			delete(clients, ws)
			break
		}
		message := clients[ws] + ": " + string(msg)
		broadcast <- []byte(message)
	}

}

func main() {
	flag.Parse()
	go handleMessages()
	http.HandleFunc("/ws", handleConnections)
	log.Println("http server started on " + *host + ":" + *port)
	err := http.ListenAndServe(":"+*port, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
