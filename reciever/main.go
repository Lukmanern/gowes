package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	defer log.Println("\n\n\n--- receiver end ---")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	endpointURL := "ws://localhost:9001"

	dialer := websocket.Dialer{}
	var c *websocket.Conn
	var dialErr error

	for c == nil {
		c, _, dialErr = dialer.Dial(endpointURL, nil)
		if dialErr != nil {
			log.Printf("Dial failed: %v. \nRetrying in 5 seconds...\n", dialErr)
			time.Sleep(5 * time.Second) // Wait and retry
		}
		if dialErr == nil {
			break
		}
	}

	defer c.Close()

	done := make(chan struct{})
	go func() {
		defer c.Close()
		defer close(done)

		for {
			_, message, readMsgErr := c.ReadMessage()
			if readMsgErr != nil {
				log.Println("read message error: " + readMsgErr.Error())
				log.Println("Retrying in 5 seconds. CTRL+C to quit")
				time.Sleep(5 * time.Second)
			}

			if readMsgErr == nil {
				log.Println("Received: " + string(message))
			}

			time.Sleep(280 * time.Millisecond)
		}
	}()

	<-interrupt

	log.Println("Interrupting")
	if wrtMsgErr := c.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	); wrtMsgErr != nil {
		log.Printf("error closing: %v\n", wrtMsgErr)
		return
	}
	<-done
	c.Close()
}
