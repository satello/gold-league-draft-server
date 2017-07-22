package main

import (
  "time"
  "encoding/json"
  "log"
)

type NominationCycle struct {
  // message channel for new nominations
  nominationChan chan *Nomination
}

func newNominationCycle() *NominationCycle {

	return &NominationCycle{
    nominationChan: make(chan *Nomination),
	}
}

// use as go routine. has callback to hub
func (d *NominationCycle) getNominee(h *DraftHub) {
  ticks := 31
  nominationTicker := time.NewTicker(time.Second)

  for {
    select {
    case <- nominationTicker.C:
      ticks -= 1
      updateCountdown(ticks, h)
      if ticks < 1 {
        nominationTicker.Stop()
        // TODO handle person not nominating someone in time
        h.startBidding <- &Player{
          Name: "shit stain",
        }
        return
      }
    case nomination := <- d.nominationChan:
      nominationTicker.Stop()
      currentPlayer := nomination.player

      currentPlayer.Bid = 1
      currentPlayer.bidderId = nomination.bidderId
      // call back to hub that you have a new player up for bid
      h.startBidding <- currentPlayer
      return
    }
  }
}

func updateCountdown(ticks int, h *DraftHub) {
  response := Response{"TICKER_UPDATE", map[string]interface{}{"ticks": ticks}}
  response_json, err := json.Marshal(response)
  if err != nil {
    log.Printf("error: %v", err)
    return
  }
  broadcastMessage(h, response_json)
}
