package maelstrom

import "encoding/json"

type messageType string

type MessageBody struct {
	// Type identifies the message type and is used to route the message to the
	// correct handler.
	Type messageType `json:"type"`
	// MessageID is a unique identifier for the message. It is used to correlate
	// requests and responses.
	MessageID int `json:"msg_id,omitempty"`
	// InReplyTo is the message ID of the message that this message is a
	// response to.
	InReplyTo int `json:"in_reply_to,omitempty"`
}

type InitBody struct {
	MessageBody
	NodeID string   `json:"node_id,omitempty"`
	Nodes  []string `json:"node_ids,omitempty"`
}

type Message struct {
	// Src refers to the ID of the node that the message originated from.
	Src string `json:"src"`
	// Dst refers to the ID of the node that the message is destined for.
	Dst string `json:"dest"`
	// Body is the body of the message. It is a json.RawMessage so the handler
	// can unmarshal it into the appropriate type.
	Body json.RawMessage `json:"body"`
}
