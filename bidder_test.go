package main

import (
  "testing"
)

func TestBidderFactoryGoodBidder(t *testing.T) {
    testName := "test"
    testCap := 54
    testSpots := 5
    goodBidder := newBidder(testName, testCap, testSpots)

    if (goodBidder.Name != testName) {
        t.Fatal("Bad Name")
        t.FailNow()
    }
    if (goodBidder.Cap != testCap) {
        t.Fatal("Bad Cap")
        t.FailNow()
    }
    if (goodBidder.Spots != testSpots) {
        t.Fatal("Bad Spots")
        t.FailNow()
    }
    if (goodBidder.BidderId == "") {
        t.Fatal("Did not generate bidderId")
        t.FailNow()
    }
    if (!goodBidder.Draftable) {
        t.Fatal("Bidder marked as not draftable")
        t.FailNow()
    }
}

func TestBidderFactoryNoCap(t *testing.T) {
    testName := "test"
    testCap := 0
    testSpots := 5
    goodBidder := newBidder(testName, testCap, testSpots)

    if (goodBidder.Draftable) {
        t.Fatal("Bidder should not be marked as draftable")
        t.FailNow()
    }
}

func TestBidderFactoryNoSpots(t *testing.T) {
    testName := "test"
    testCap := 25
    testSpots := 0
    goodBidder := newBidder(testName, testCap, testSpots)

    if (goodBidder.Draftable) {
        t.Fatal("Bidder should not be marked as draftable")
        t.FailNow()
    }
}
