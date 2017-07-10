package main

import (
  "fmt"
  "encoding/json"
  "log"
)

type Response struct {
  MessageType string
  Body map[string]interface{}
}

// draft hub maintains the set of active clients and broadcasts messages to the
// clients.
type DraftHub struct {
	// Registered clients.
	clients map[*Subscriber]bool

	// Inbound messages from the clients.
	broadcast chan []byte

	// Register requests from the clients.
	register chan *Subscriber

	// Unregister requests from clients.
	unregister chan *Subscriber

  // accept message from client
  acceptMessage chan *Message

  // players eligable for draft
  players map[string]*Player

  // bidders in the draft
  bidders map[string]*Bidder
}

func newDraft() *DraftHub {
	return &DraftHub{
		broadcast:      make(chan []byte),
		register:       make(chan *Subscriber),
		unregister:     make(chan *Subscriber),
    acceptMessage:  make(chan *Message),
		clients:        make(map[*Subscriber]bool),
    players:        make(map[string]*Player),
    bidders:        make(map[string]*Bidder),
	}
}

func broadcastMessage(h *DraftHub, message []byte) {
  for client := range h.clients {
    select {
    case client.send <- message:
    default:
      close(client.send)
      delete(h.clients, client)
    }
  }
}

func (h *DraftHub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case messageJson := <-h.acceptMessage:
      switch t := messageJson.MessageType; t {
    	case "newBidder":
        fmt.Println("new bidder")
    	case "chatMessage":
        body := messageJson.Body

        response := Response{"CHAT_MESSAGE", body}
        response_json, err := json.Marshal(response)
        if err != nil {
    			log.Printf("error: %v", err)
    			break
        }
        broadcastMessage(h, []byte(response_json))
    	default:
    		// freebsd, openbsd,
    		// plan9, windows...
    		fmt.Printf("%s.", t)
      }
		}
	}
}
