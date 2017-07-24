package main

import (
  "log"
  "encoding/json"
)

type Nomination struct {
  // player nominated
  player *Player

  // id of the bidder that nominated player
  bidderId string
}

type Bid struct {
  // amount of the bid
  amount int

  // timestamp of bid
  // TODO do I want this?

  // who made the bid
  bidderId string
}

func broadcastNewBidderNominee(bidder *Bidder, h *DraftHub) {
  response := Response{"NEW_NOMINEE", map[string]interface{}{"bidderId": bidder.BidderId}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  log.Printf("%s", response_json)
  broadcastMessage(h, response_json)
}

func broadcastNewPlayerNominee(player *Player, h *DraftHub) {
  response := Response{"NEW_PLAYER_NOMINEE", map[string]interface{}{"name": player.Name}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  log.Printf("%s", response_json)
  broadcastMessage(h, response_json)
}
