package main

import (
  "time"
  "encoding/json"
  "log"
)

type DraftLoop struct {
  // draft hub using loop
  draftHub *DraftHub

  // draft index
  draftIndex int

  // list of Bidders in order of nomination
  nominationOrder []*Bidder
}

func newDraftLoop(bidders []*Bidder, h *DraftHub) *DraftLoop {

	return &DraftLoop{
		draftHub: h,
    draftIndex: 0,
    nominationOrder: bidders,
	}
}

func (d *DraftLoop) start() {
  for {
    // stop once there is nobody left to draft
    if len(d.nominationOrder) < 1 {
      break
    }

    ticker := time.NewTicker(time.Second)
    ticks := 30
    go func() {
        for range ticker.C {
          ticks -= 1
          updateCountdown(ticks, d.draftHub)
        }
    }()
    time.Sleep(time.Second * 30)
    ticker.Stop()

    broadcastNewBidderNominee(d.nominationOrder[d.draftIndex], d.draftHub)
    d.draftIndex = (d.draftIndex + 1) % len(d.nominationOrder)
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
