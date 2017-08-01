package main

import (
  "time"
  "log"
)

type NominationCycle struct {
  // message channel for new nominations
  nominationChan chan *Nomination

  // pause chan for telling cycle to pause
  pauseChan chan bool

  // channel for interupting cycle
  interuptChan chan bool

  // bool indicating if open
  open bool
}

func newNominationCycle() *NominationCycle {

	return &NominationCycle{
    nominationChan: make(chan *Nomination),
    pauseChan: make(chan bool),
    interuptChan: make(chan bool),
    open: false,
	}
}

// use as go routine. has callback to hub
func (d *NominationCycle) getNominee(h *DraftHub, bidderId string) {
  d.open = true
  ticks := 30
  updateCountdown(ticks, h)
  nominationTicker := time.NewTicker(time.Second)

  loop:
  for {
    select {
    case <- nominationTicker.C:
      ticks -= 1
      updateCountdown(ticks, h)
      if ticks < 1 {
        nominationTicker.Stop()
        // TODO handle person not nominating someone in time
        autoPlayer := h.players.getHighestValuePlayer()
        // set auto nomination
        h.draftState.CurrentBid = 1
        h.draftState.CurrentBidderId = bidderId
        h.draftState.CurrentPlayerName = autoPlayer.Name
        h.startBidding <- autoPlayer
        autoPlayer.bid = &Bid{
          amount: 1,
          bidderId: bidderId,
        }
        d.open = false
        break loop
      }
    case nomination := <- d.nominationChan:
      nominationTicker.Stop()
      currentPlayer := nomination.player
      var amount int
      amount = nomination.amount
      if amount <= 0 {
        amount = 1
      }

      currentPlayer.bid = &Bid{
        amount: amount,
        bidderId: bidderId,
      }

      // update draft state
      h.draftState.CurrentBid = amount
      h.draftState.CurrentBidderId = bidderId
      h.draftState.CurrentPlayerName = currentPlayer.Name
      // call back to hub that you have a new player up for bid
      h.startBidding <- currentPlayer
      d.open = false
      break loop
    case shouldPause := <- d.pauseChan:
      if shouldPause {
        nominationTicker.Stop()
      } else {
        // just to make sure last one is stopped
        nominationTicker.Stop()
        nominationTicker = time.NewTicker(time.Second)
      }
    case interupt := <- d.interuptChan:
      log.Println("stopping....")
      if interupt {
        d.open = false
        break loop
      }
    }
  }
}
