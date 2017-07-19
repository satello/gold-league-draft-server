package main

import (
	"log"
	"encoding/json"
)

type Player struct {
  // player name
	Name string `json:"name"`

  // player position
	Position string `json:"position"`

	// player bye week
	Bye int `json:"bye"`

	// arbitrary value metric
	Value int `json:"value"`

  // current bid
  Bid int `json:"bid"`

  // bid owner
  bidder Bidder
}

func newPlayer(name string, position string) *Player {
  return &Player{
    Name:     name,
    Position: position,
    Bid:      0,
  }
}

func (p *Player) submitBid(bid int, bidder Bidder) bool {
  if bid > p.Bid && p.bidder != bidder {
    p.Bid = bid
    p.bidder = bidder
    return true
  } else {
    return false
  }
}

func getPlayers(s *Subscriber, h *DraftHub) {
  log.Printf("GET PLAYERS")
	// FIXME this is kind of dumb
	var playerSlice []*Player
  for _, v := range h.players {
    playerSlice = append(playerSlice, v)
  }
	log.Printf("number of players in slice %d", len(playerSlice))

  response := Response{"GET_PLAYERS", map[string]interface{}{"players": playerSlice}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  sendMessageToSubscriber(h, s, response_json)
}
