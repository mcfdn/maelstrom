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
./maelstrom/maelstrom test -w echo --bin main --nodes n1 --time-limit 10
```
