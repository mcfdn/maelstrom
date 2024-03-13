# Maelstrom

This project contains solutions for the
[Maelstrom](https://github.com/jepsen-io/maelstrom/blob/main/README.md)
distributed systems workbench exercises.

All solutions are written in the Go programming language (version 1.22.1).

# Solutions

## Echo

Reference: https://github.com/jepsen-io/maelstrom/blob/main/doc/02-echo/index.md

```
go build -o echo cmd/echo/main.go
./maelstrom test -w echo --bin echo --nodes n1 --time-limit 10
```

## Broadcast

Reference: https://github.com/jepsen-io/maelstrom/blob/main/doc/03-broadcast/01-broadcast.md

```
go build -o broadcast cmd/broadcast/main.go
./maelstrom test -w broadcast --bin broadcast --time-limit 5 --rate 10
```
