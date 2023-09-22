package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var clients = make(map[*websocket.Conn]bool)
var clientMutex sync.Mutex

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrading error: %v\n", err)
		return
	}
	defer func() {
		log.Println("\n\n --- ECHO Reciever END ---")
		c.Close()
	}()

	clientMutex.Lock()
	clients[c] = true
	clientMutex.Unlock()

	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("Reading error: %v\n", err)
			clientMutex.Lock()
			delete(clients, c)
			clientMutex.Unlock()
			break
		}

		log.Printf("recv: message %q", message)

		clientMutex.Lock()
		// Broadcast the received message to all connected clients
		for client := range clients {
			if err := client.WriteMessage(mt, message); err != nil {
				log.Printf("Writing error: %v\n", err)
				client.Close()
				delete(clients, client)
			}
		}
		clientMutex.Unlock()
	}
}

func main() {
	p := 9001
	log.Printf("Starting websocket echo server on port %d", p)
	defer log.Println("\n\n\n--- server end ---")

	http.HandleFunc("/", echo)

	go func() {
		for {
			fmt.Print("Enter a message to broadcast (or 'quit' to exit): ")
			scanner := bufio.NewScanner(os.Stdin)
			if scanner.Scan() {
				message := scanner.Text()

				if message == "quit" {
					os.Exit(0)
				}

				clientMutex.Lock()
				// Broadcast the message to all connected clients
				for client := range clients {
					if err := client.WriteMessage(websocket.TextMessage, []byte(message)); err != nil {
						log.Printf("Writing error: %v\n", err)
						client.Close()
						delete(clients, client)
					}
				}
				clientMutex.Unlock()
			}
		}
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", p), nil); err != nil {
		log.Panicf("Error while starting to listen: %v", err)
	}
}
