package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/websocket"
)

type Output struct {
	Info  *log.Logger
	Error *log.Logger
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(*http.Request) bool { return true },
}
var registry = ChannelRegistry{}

func main() {
	output := Output{
		Info:  log.New(os.Stdout, "[go-sock-yourself] ", log.Ldate|log.Ltime),
		Error: log.New(os.Stderr, "[go-sock-yourself] ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	queue := AMQPListener{}

	queue.SetURL("amqp://eyy:nope@queue.expandonline.nl:5672/staging")
	queue.SetExchange("test")
	queue.SetOutput(&output)

	output.Info.Println("Starting AQMP listener")
	go queue.Init(&registry)

	http.HandleFunc("/echo", echo)

	output.Info.Println("Starting HTTP server")
	log.Fatal(http.ListenAndServe("localhost:4321", nil))
}

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("Upgrade:", err)
		return
	}

	defer c.Close()

	for {
		mt, userID, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		channel := registry.Ensure(string(userID))
		response := <-channel

		err = c.WriteMessage(mt, []byte(response.toBytes()))
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
