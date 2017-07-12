package main

import (
  "encoding/json"
  "log"
  "os/exec"
)

type Bidder struct {
  // name
  Name string `json:"name"`

  // cap room
  Cap int     `json:"cap"`

  // roster spots
  Spots int   `json:"spots"`
}

func newBidder(name string, cap int, spots int) *Bidder {
  return &Bidder{
    Name: name,
    Cap: cap,
    Spots: spots,
  }
}

func createBidderId() string {
  // return UUID
  out, err := exec.Command("uuidgen").Output()
  if err != nil {
      log.Fatal(err)
  }
  return string(out[:])
}

func createBidder(name string, cap int, spots int, s *Subscriber, h *DraftHub) {
  log.Println("NEW BIDDER")
  new_bidder := newBidder(name, cap, spots)

  // create token for bidder. use token as key
  token := createBidderId()
  h.bidders[token] = new_bidder

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
  b := h.bidders[token]
  if b != nil {
    response := Response{"TOKEN_VALID", nil}
    response_json, err := json.Marshal(response)
    if err != nil {
      log.Printf("error: %v", err)
      return
    }
    log.Println(string(response_json))
    // attach bidderId to connection
    s.bidderId = token
    sendMessageToSubscriber(h, s, response_json)
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
  if _, ok := h.bidders[token]; ok {
    delete(h.bidders, token)
  }

  s.bidderId = ""
}

func getBidders(s *Subscriber, h *DraftHub) {
  var bidderSlice []*Bidder
  for _, v := range h.bidders {
    bidderSlice = append(bidderSlice, v)
    r, _ := json.Marshal(v)
    log.Printf("%s", r)
  }

  log.Println(h.bidders)
  log.Println(bidderSlice)

  response := Response{"GET_BIDDERS", map[string]interface{}{"bidders": bidderSlice}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  log.Printf("%s", response_json)
  sendMessageToSubscriber(h, s, response_json)
}
