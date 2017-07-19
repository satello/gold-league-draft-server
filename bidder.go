package main

import (
  "encoding/json"
  "log"
)

type Bidder struct {
  // name
  Name string `json:"name"`

  // cap room
  Cap int     `json:"cap"`

  // roster spots
  Spots int   `json:"spots"`

  // bidder uuid
  BidderId string `json:"bidderId"`

  // has an active websocket connection
  ActiveConnection bool `json:"activeConnection"`
}

func newBidder(name string, cap int, spots int) *Bidder {
  return &Bidder{
    Name: name,
    Cap: cap,
    Spots: spots,
  }
}

func createBidder(name string, cap int, spots int, s *Subscriber, h *DraftHub) {
  log.Println("NEW BIDDER")
  new_bidder := newBidder(name, cap, spots)

  // create token for bidder. use token as key
  token := createUuid()
  h.biddersMap[token] = new_bidder

  token_json := map[string]interface{}{"token": token}
  response := Response{"NEW_TOKEN", token_json}
  response_json, err := json.Marshal(response)
  log.Println(string(response_json))
  if err != nil {
    log.Printf("error: %v", err)
    return
  }

  // attach bidderId to connection
  s.bidderId = token
  sendMessageToSubscriber(h, s, response_json)
}

func authorizeBidder(token string, s *Subscriber, h *DraftHub) {
  log.Println("AUTHORIZE BIDDER")
  b := h.biddersMap[token]
  if b != nil {
    // bidder does not have an active connection
    if !b.ActiveConnection {
      response := Response{"TOKEN_VALID", nil}
      response_json, err := json.Marshal(response)
      if err != nil {
        log.Printf("error: %v", err)
        return
      }
      log.Println(string(response_json))
      // attach bidderId to connection
      s.bidderId = token
      // mark connection as active
      b.ActiveConnection = true
      sendMessageToSubscriber(h, s, response_json)
    } else {
      response := Response{"DUPLICATE_CONNECTION", nil}
      response_json, err := json.Marshal(response)
      if err != nil {
        log.Printf("error: %v", err)
        return
      }
      log.Println(string(response_json))
      sendMessageToSubscriber(h, s, response_json)
    }
  } else {
    response := Response{"INVALID_TOKEN", nil}
    response_json, err := json.Marshal(response)
    if err != nil {
      log.Printf("error: %v", err)
      return
    }
    log.Println(string(response_json))
    sendMessageToSubscriber(h, s, response_json)
  }
}

func deactivateBidder(token string, s *Subscriber, h *DraftHub) {
  log.Printf("DEAUTHORIZE BIDDER")
  if _, ok := h.biddersMap[token]; ok {
    delete(h.biddersMap, token)
  }

  s.bidderId = ""
}

func getBidders(s *Subscriber, h *DraftHub) {
  log.Printf("GET BIDDERS")

  response := Response{"GET_BIDDERS", map[string]interface{}{"bidders": h.biddersSlice}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  log.Printf("%s", response_json)
  sendMessageToSubscriber(h, s, response_json)
}
