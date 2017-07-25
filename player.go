package main

import (
	"log"
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
  bid *Bid
}

func newPlayer(name string, position string) *Player {
  return &Player{
    Name:     name,
    Position: position,
    bid:      &Bid{
			amount: 0,
		},
  }
}

// func (p *Player) submitBid(bid int, bidder Bidder) bool {
//   if bid > p.Bid && p.bidder != bidder {
//     p.Bid = bid
//     p.bidder = bidder
//     return true
//   } else {
//     return false
//   }
// }

func getPlayers(s *Subscriber, h *DraftHub) {
  log.Printf("GET PLAYERS")
	// FIXME this is kind of dumb
	var playerSlice []*Player
  for _, v := range h.players {
    playerSlice = append(playerSlice, v)
  }
	log.Printf("number of players in slice %d", len(playerSlice))

  response := Response{"GET_PLAYERS", map[string]interface{}{"players": playerSlice}}
  response_json := responseToJson(response)
  sendMessageToSubscriber(h, s, response_json)
}
