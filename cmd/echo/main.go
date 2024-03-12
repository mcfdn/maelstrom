package main

import (
	"encoding/json"
	"fmt"

	"github.com/mcfdn/maelstrom"
)

type EchoBody struct {
	maelstrom.MessageBody
	Echo string `json:"echo"`
}

func main() {
	node := maelstrom.NewNode()
	err := node.Handle("echo", echoHandler(node))
	if err != nil {
		panic(err)
	}
	err = node.Run()
	if err != nil {
		panic(err)
	}
}

func echoHandler(node *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body EchoBody
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return fmt.Errorf("error unmarshalling message: %v", err)
		}
		body.Type = "echo_ok"
		body.InReplyTo = body.MessageID
		body.MessageID = 0
		return node.Send(msg.Src, body)
	}
}
