package main

type Player struct {
  // player name
	Name string

  // player position
	Position string

  // current bid
  Bid int

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
