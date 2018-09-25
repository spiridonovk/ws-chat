package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
)

var msgCh = make(chan []byte)
var addr = flag.String("addr", "localhost:8000", "http service address")

func getNick(reader *bufio.Reader) (string, error) {
	fmt.Print("Enter nickname: ")
	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if len(text) < 2 {
			log.Print("Nick cann't be empty")
			continue
		}
		return text, nil
	}
}

func sendMesage(c *websocket.Conn) {
	for {
		msg := <-msgCh
		err := c.WriteMessage(1, msg)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func main() {
	flag.Parse()
	reader := bufio.NewReader(os.Stdin)
	nickName, err := getNick(reader)
	if err != nil {
		log.Printf("ReadString error %s: ", err)
		return
	}
	u := url.URL{Scheme: "ws", Host: *addr, Path: "/ws"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"nick": {nickName}})
	if err != nil {
		log.Fatal("dial:", err)
	}
	go sendMesage(c)
	defer c.Close()

	go func() {
		for {
			_, msg, err := c.ReadMessage()
			if err != nil {
				log.Printf("ReadMessage error:  %v", err.Error())
				return
			}
			fmt.Printf("%v \n", string(msg))
		}
	}()

	go func(reader *bufio.Reader) {
		for {
			text, _ := reader.ReadString('\n')
			msgCh <- []byte(text)
		}
	}(reader)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-interrupt
	log.Printf("received signal, exiting: %v", sig)
	log.Print("goodbye")
	os.Exit(0)

}
