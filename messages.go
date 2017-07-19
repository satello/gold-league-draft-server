package main

// WS MESSGES FROM FRONT END
type Message struct {
    MessageType string
    BidderId string
    Body map[string]interface{}
    Subscriber *Subscriber
}

type NewBidderBody struct {
  Name string `json:"name"`
  Cap int     `json:"cap"`
  Spots int   `json:"space"`
}

type TokenBody struct {
  Token string `json:"token"`
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
