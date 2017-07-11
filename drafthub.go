package main

import (
  "fmt"
  "encoding/json"
  "log"
  "os/exec"
  "github.com/mitchellh/mapstructure"
)

type Response struct {
  MessageType string
  Body map[string]interface{} `json:"body"`
}

type NewBidderBody struct {
  name string
  cap int
  spots int
}

type AuthorizeTokenBody struct {
  Token string `json:"token"`
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

// SEND MESSAGE TO ALL CLIENTS
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

// SEND MESSAGE TO A SINGLE CLIENT
func sendMessageToSubscriber(h *DraftHub, c *Subscriber, message []byte) {
  select {
  case c.send <- message:
  default:
    close(c.send)
    delete(h.clients, c)
  }
}

func createToken() string {
  // return UUID
  out, err := exec.Command("uuidgen").Output()
  if err != nil {
      log.Fatal(err)
  }
  return string(out[:])
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
        log.Println("NEW BIDDER")
        var body NewBidderBody
        mapstructure.Decode(messageJson.Body, &body)
        name := body.name
        cap := body.cap
        spots := body.spots
        new_bidder := Bidder{name, cap, spots}
        // create token for bidder. use token as key
        token := createToken()
        h.bidders[token] = &new_bidder

        token_json := map[string]interface{}{"token": token}
        response := Response{"NEW_TOKEN", token_json}
        response_json, err := json.Marshal(response)
        log.Println(string(response_json))
        if err != nil {
    			log.Printf("error: %v", err)
    			break
        }

        sendMessageToSubscriber(h, messageJson.Subscriber, response_json)

      case "authorizeToken":
        log.Printf("AUTH TOKEN");
        var body AuthorizeTokenBody
        mapstructure.Decode(messageJson.Body, &body)

        token := body.Token
        s := h.bidders[token]
        if s != nil {
          response := Response{"TOKEN_VALID", nil}
          response_json, err := json.Marshal(response)
          if err != nil {
      			log.Printf("error: %v", err)
      			break
          }
          log.Println(string(response_json))
          sendMessageToSubscriber(h, messageJson.Subscriber, response_json)
        } else {
          response := Response{"INVALID_TOKEN", nil}
          response_json, err := json.Marshal(response)
          if err != nil {
      			log.Printf("error: %v", err)
      			break
          }
          log.Println(string(response_json))
          sendMessageToSubscriber(h, messageJson.Subscriber, response_json)
        }

    	case "chatMessage":
        log.Printf("CHAT MESSAGE");
        body := messageJson.Body

        response := Response{"CHAT_MESSAGE", body}
        response_json, err := json.Marshal(response)
        if err != nil {
    			log.Printf("error: %v", err)
    			break
        }
        broadcastMessage(h, response_json)

    	default:
    		// freebsd, openbsd,
    		// plan9, windows...
    		fmt.Printf("%s.", t)
      }
		}
	}
}
