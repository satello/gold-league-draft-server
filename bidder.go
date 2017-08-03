package main

import (
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

  // whether the bidder is eligable to keep drafting
  Draftable bool `json:"eligible"`
}

func newBidder(name string, cap int, spots int) *Bidder {
  // determine if bidder is draft eligable
  var draftable bool
  if cap < 1 || spots < 1 {
    draftable = false
  } else {
    draftable = true
  }

  // create bidder id
  bidderId := createUuid()

  return &Bidder{
    Name: name,
    Cap: cap,
    Spots: spots,
    BidderId: bidderId,
    Draftable: draftable,
  }
}

func authorizeBidder(token string, s *Subscriber, h *DraftHub) {
  log.Println("AUTHORIZE BIDDER")
  b := h.biddersMap[token]
  if b != nil {
    // bidder does not have an active connection
    if !b.ActiveConnection {
      response := Response{"TOKEN_VALID", nil}
      response_json := responseToJson(response)
      // attach bidderId to connection
      s.bidderId = token
      // mark connection as active
      log.Println("ACTIVATING CONNECTION")
      b.ActiveConnection = true
      // send response to subscriber
      sendMessageToSubscriber(h, s, response_json)
      // braodcast that there is a new bidder
      broadcastBidderState(b, h)
    } else {
      response := Response{"INVALID_TOKEN", nil}
      response_json := responseToJson(response)
      sendMessageToSubscriber(h, s, response_json)
    }
  } else {
    response := Response{"INVALID_TOKEN", nil}
    response_json := responseToJson(response)
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
  response_json := responseToJson(response)
  sendMessageToSubscriber(h, s, response_json)
}

func broadcastBidderState(b *Bidder, h *DraftHub) {
  log.Printf("BIDDER %s STATE CHANGE", b.Name)

  response := Response{"BIDDER_STATE_CHANGE", map[string]interface{}{"bidder": b}}
  response_json := responseToJson(response)
  broadcastMessage(h, response_json)
}
