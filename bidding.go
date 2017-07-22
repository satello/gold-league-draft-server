package main

import (
  "log"
  "encoding/json"
)

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
