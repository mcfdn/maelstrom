package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/mcfdn/maelstrom"
)

var messages []int
var mu sync.RWMutex

type TopologyBody struct {
	maelstrom.MessageBody
	Topology map[string][]string `json:"topology"`
}

type ReadBody struct {
	maelstrom.MessageBody
	Messages []int `json:"messages"`
}

type BroadcastBody struct {
	maelstrom.MessageBody
	Message int `json:"message"`
}

func main() {
	node := maelstrom.NewNode()
	err := node.Handle("topology", topologyHandler(node))
	if err != nil {
		panic(err)
	}
	err = node.Handle("read", readHandler(node))
	if err != nil {
		panic(err)
	}
	err = node.Handle("broadcast", broadcastHandler(node))
	if err != nil {
		panic(err)
	}
	_ = node.Handle("broadcast_ok", func(msg maelstrom.Message) error {
		// Avoid an unknown handler error; no response is required.
		return nil
	})
	err = node.Run()
	if err != nil {
		panic(err)
	}
}

func topologyHandler(node *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		var body TopologyBody
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return fmt.Errorf("error unmarshalling message: %v", err)
		}
		if body.Topology[node.ID()] == nil {
			return fmt.Errorf("node %s not in topology", node.ID())
		}
		node.SetTopology(body.Topology[node.ID()])
		reply := maelstrom.MessageBody{
			Type:      "topology_ok",
			InReplyTo: body.MessageID,
		}
		return node.Send(msg.Src, reply)
	}
}

func readHandler(node *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		mu.RLock()
		defer mu.RUnlock()

		var body maelstrom.MessageBody
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return fmt.Errorf("error unmarshalling message: %v", err)
		}
		reply := ReadBody{
			Messages: messages,
			MessageBody: maelstrom.MessageBody{
				Type:      "read_ok",
				InReplyTo: body.MessageID,
			},
		}
		return node.Send(msg.Src, reply)
	}
}

func broadcastHandler(node *maelstrom.Node) func(maelstrom.Message) error {
	return func(msg maelstrom.Message) error {
		mu.Lock()
		defer mu.Unlock()

		var body BroadcastBody
		err := json.Unmarshal(msg.Body, &body)
		if err != nil {
			return fmt.Errorf("error unmarshalling message: %v", err)
		}
		reply := maelstrom.MessageBody{
			Type:      "broadcast_ok",
			InReplyTo: body.MessageID,
		}
		if messageIsKnown(body.Message) {
			return node.Send(msg.Src, reply)
		}

		messages = append(messages, body.Message)
		for _, n := range node.Topology() {
			if n == msg.Src {
				// No need to send back to the sender.
				continue
			}
			node.Send(n, BroadcastBody{
				MessageBody: maelstrom.MessageBody{
					Type: "broadcast",
				},
				Message: body.Message,
			})
		}
		return node.Send(msg.Src, reply)
	}
}

func messageIsKnown(message int) bool {
	for _, m := range messages {
		if m == message {
			return true
		}
	}
	return false
}
