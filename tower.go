package main

import (
  "log"
)

type Tower struct {
  rooms map[string]*DraftHub
}

func newTower() *Tower {
  return &Tower{
    rooms: make(map[string]*DraftHub),
  }
}

type Room struct {
  RoomId string `json:"roomId"`

  Rules *Rules `json:"rules"`

  Bidders []*Bidder `json:"bidders"`

  Players []*Player `json:"players"`
}

func newRoom(t *Tower, rules *Rules, bidders []*Bidder, players []*Player) string {
  log.Println("starting new room")
  log.Println(len(bidders))
  roomId := createUuid()
  newDraftRoom := newDraft(rules, bidders, players)

  // start new hub
  go newDraftRoom.run()

  // TODO watch out or memory leaks with this. Do go routines shut down when the parent does?
  log.Println("new draft room created")
  log.Println(roomId)
  t.rooms[roomId] = newDraftRoom
  log.Println(len(t.rooms))
  log.Println(t.rooms[roomId])

  return roomId
}
