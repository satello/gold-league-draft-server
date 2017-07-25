package main

import (
  "time"
  "log"
)

type BiddingCycle struct {
  // message channel for new nominations
  biddingChan chan *Bid

  open bool
}

func newBiddingCycle() *BiddingCycle {

	return &BiddingCycle{
    biddingChan: make(chan *Bid),
    open: false,
	}
}

// use as go routine. has callback to hub
func (d *BiddingCycle) getBids(player *Player, h *DraftHub) {
  d.open = true
  ticks := 30
  updateCountdown(ticks, h)
  biddingTicker := time.NewTicker(time.Second)

  loop:
  for {
    select {
    case <- biddingTicker.C:
      ticks -= 1
      updateCountdown(ticks, h)
      if ticks < 1 {
        biddingTicker.Stop()

        h.endBidding <- player
        d.open = false
        break loop
      }
    case bid := <- d.biddingChan:
      currentBid := player.bid.amount
      // currentBidderId := player.bid.bidderId

      // skip bid if owner already has top bid or bid isn't high enough
      if bid.amount <= currentBid {
        continue
      }

      player.bid = bid
      log.Println(bid)
      // reset timer if ticks < 12 seconds
      if ticks < 15 {
        ticks = 15
        updateCountdown(ticks, h)
      }
      broadcastNewPlayerBid(player, h)
    }
  }
}
