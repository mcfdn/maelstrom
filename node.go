package maelstrom

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// initHandler describes a function signature for handling the `init` message.
type initHandler func(node *Node, msg Message) error

// messageHandler describes a function signature for handling messages sent to
// a node.
type messageHandler func(msg Message) error

// defaultInitHandler is the default handler for the `init` message. It reads the
// message body and initialises the provided node with the ID and list of nodes
// in the cluster. It then sends an `init_ok` message back to the source of the
// `init` message.
func defaultInitHandler(node *Node, msg Message) error {
	var body InitBody
	err := json.Unmarshal(msg.Body, &body)
	if err != nil {
		return err
	}
	node.Init(body.NodeID, body.Nodes)

	reply := MessageBody{
		Type:      "init_ok",
		InReplyTo: body.MessageID,
	}
	return node.Send(msg.Src, reply)
}

// Node represents a node in the network.
type Node struct {
	// reader is the input stream to read messages from.
	reader io.Reader
	// writer is the output stream to write messages to.
	writer io.Writer
	// handlers is a map of message types to message handlers.
	handlers map[messageType]messageHandler
	// id is the uniqute ID of this node in the cluster.
	id string
	// nodes is the list of node IDs in the cluster, including the local node.
	nodes []string
	// topology is the list of node IDs in the neighbourhood of the local node.
	topology []string
	// initHandler is the handler for the `init` message. It uses a different
	// message signature to regular message handlers, and is expected to be
	// invoked before any other messages are received.
	initHandler initHandler
}

// NewNode allocates and returns a new Node.
func NewNode() *Node {
	return &Node{
		reader:      io.Reader(os.Stdin),
		writer:      io.Writer(os.Stdout),
		handlers:    make(map[messageType]messageHandler),
		initHandler: defaultInitHandler,
	}
}

// Init initialises the node with the id and list of nodes in the cluster. This
// information is received in an `init` message.
func (n *Node) Init(id string, nodes []string) {
	n.id = id
	n.nodes = nodes
}

func (n *Node) ID() string {
	return n.id
}

func (n *Node) Topology() []string {
	return n.topology
}

func (n *Node) SetTopology(topology []string) {
	n.topology = topology
}

// Handle registers handler against the given msgType. If handler already exists
// for the msgType, an error is returned.
func (n *Node) Handle(msgType messageType, handler messageHandler) error {
	if msgType == "init" {
		return fmt.Errorf("cannot register handler for init message type")
	}
	if _, ok := n.handlers[msgType]; ok || msgType == "init" {
		return fmt.Errorf("handler for %s already exists", msgType)
	}
	n.handlers[msgType] = handler
	return nil
}

// Send sends the given body to the provided dst node by marshalling the body to
// JSON and writing to writer.
func (n *Node) Send(dst string, body any) error {
	b, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("error marshalling message body: %w", err)
	}
	msg := Message{
		Src:  n.id,
		Dst:  dst,
		Body: b,
	}
	b, err = json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("error marshalling message: %w", err)
	}
	b = append(b, '\n')
	_, err = n.writer.Write(b)
	if err != nil {
		return fmt.Errorf("error writing message to writer: %w", err)
	}
	return nil
}

// Run starts the node. It reads newline-delimited messages from the reader and
// dispatches them to the appropriate handler. This function blocks until the
// read stream is closed (EOF). Any errors encountered during message handling
// are returned, ending the read loop.
func (n *Node) Run() error {
	scanner := bufio.NewScanner(n.reader)
	for scanner.Scan() {
		var msg Message
		err := json.Unmarshal(scanner.Bytes(), &msg)
		if err != nil {
			return fmt.Errorf("error unmarshalling message: %w", err)
		}
		var body MessageBody
		err = json.Unmarshal(msg.Body, &body)
		if err != nil {
			return fmt.Errorf("error unmarshalling message body: %w", err)
		}

		// Handle the `init` message separately here, since each node needs to
		// be initialised before any other messages are handled.
		if body.Type == "init" {
			err = n.initHandler(n, msg)
			if err != nil {
				return fmt.Errorf("error handling init message: %w", err)
			}
			continue
		}

		if !n.isInitialised() {
			return fmt.Errorf("received message of type %s before init", body.Type)
		}

		// Resolve the correct handler for this message type and invoke it.
		if handler, ok := n.handlers[body.Type]; ok {
			if err := handler(msg); err != nil {
				return fmt.Errorf(
					"error handling message of type %s: %w",
					body.Type,
					err,
				)
			}
			continue
		}
		return fmt.Errorf("no handler for message type %s", body.Type)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading from reader: %w", err)
	}
	return nil
}

func (n *Node) isInitialised() bool {
	return n.id != ""
}
