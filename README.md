A fork of https://github.com/satello/auction-server

Used in conjuction with:

https://github.com/satello/gold-league-scrape
(Service that scrapes several sources to serve information about the gold league via a REST api)

and

https://github.com/satello/gold-league-draft-room
(Front end for auction)

----

This server written in Go connects to a draft room via websockets and facilitates a live auction.

Uses gold-league-scrape to connect to google doc for reading and writing. This is just a personal project for a fantasy football league. For code more applicable to what you are working on check out the more generalized version here:
https://github.com/satello/auction-server

## Quickstart

1) Install Go and set GoPath

2) Install dependencies (FIXME create a MAKEFILE to make this nice and easy)

3)
```
go build
```

4) This should compile to a binary with the same name as your directory
```
./<compiled binary>
