package main

type Player struct {
  // player name
	name string

  // player position
	position string

  // current bid
  bid int

  // bid owner
  bidder Bidder
}

func newPlayer(name string, position string) *Player {
  return &Player{
    name:     name,
    position: position,
    bid:      0,
  }
}

func (p *Player) submitBid(bid int, bidder Bidder) bool {
  if bid > p.bid && p.bidder != bidder {
    p.bid = bid
    p.bidder = bidder
    return true
  } else {
    return false
  }
}
