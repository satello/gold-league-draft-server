package main

import (
  "time"
  "log"
)

type BiddingCycle struct {
  // message channel for new nominations
  biddingChan chan *Bid

  // pause chan for telling cycle to pause
  pauseChan chan bool

  // channel for interupting cycle
  interuptChan chan bool

  // bool indicating if open
  open bool
}

func newBiddingCycle() *BiddingCycle {

	return &BiddingCycle{
    biddingChan: make(chan *Bid),
    pauseChan: make(chan bool),
    interuptChan: make(chan bool),
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

      // skip bid if owner already has top bid or bid isn't high enough
      if bid.amount <= currentBid {
        continue
      }

      // update player state
      player.bid = bid
      // update draftState
      h.draftState.CurrentBid = bid.amount
      h.draftState.CurrentBidderId = bid.bidderId
      // reset timer if ticks < 12 seconds
      if ticks < 15 {
        ticks = 15
        updateCountdown(ticks, h)
      }
      broadcastNewPlayerBid(player, h)
    case shouldPause := <- d.pauseChan:
      if shouldPause {
        biddingTicker.Stop()
      } else {
        biddingTicker = time.NewTicker(time.Second)
      }
    case interupt := <- d.interuptChan:
      log.Println("interupt bidding cycle")
      if interupt {
        d.open = false
        break loop
      }
    }
  }
}
