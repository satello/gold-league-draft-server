package main

type Bidder struct {
  // name
  name string

  // cap room
  cap int

  // roster spots
  spotsAvailable int
}

func newBidder(name string, cap int, spotsAvailable int) *Bidder {
  return &Bidder{
    name: name,
    cap: cap,
    spotsAvailable: spotsAvailable,
  }
}
