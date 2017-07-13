package main


type Rules struct {
  // uses spots
  LimitedSpots bool `json:"limitedSpots"`

  // requires that you fill all spots
  AllSpotsFilled bool `json:"allSpotsFilled"`

  // use auctioneer or allow participants to nominate
  UseAutoNominate bool `json:"useAutoNominate"`

  // seconds per item
  MinSecondsPerItem int `json:"minSecondsPerItem"`

  // reset time on bid
  ResetTimerOnBid bool `json:"resetTimerOnBid"`

  // amount of time to reset to
  ResetSeconds int `json:"resetSeconds"`
}
