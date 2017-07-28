package main

import (
  "log"
  "time"
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

  Bidders []*Bidder `json:"bidders"`

  Players []*Player `json:"players"`
}

func newRoom(t *Tower, bidders []*Bidder, players []*Player) string {
  roomId := createUuid()
  newDraftRoom := newDraft(bidders, players, roomId)

  // start new hub
  go newDraftRoom.run()
  timer := time.NewTimer(time.Hour * 24)
  go func() {
    <- timer.C
    newDraftRoom.isActive = false
    newDraftRoom.draftState.Running = false
    delete(t.rooms, roomId)
    log.Printf("room %s no longer active", roomId)
  }()

  // TODO watch out or memory leaks with this. Do go routines shut down when the parent does?
  t.rooms[roomId] = newDraftRoom

  return roomId
}
