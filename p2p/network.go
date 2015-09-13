package p2p

import (
	"log"
	"strings"
	"time"

	"github.com/hashicorp/memberlist"
)

var (
	printMessages = true
)

func PrintMessages(b bool) {
	printMessages = b
}

type Config *memberlist.Config

type Callback func(peer string, messageType, message string)
type Network struct {
	Memberlist  *memberlist.Memberlist
	Subscribers map[string][]Callback
}

func DefaultConfig() Config {
	return Config(memberlist.DefaultLANConfig())
}

func New(cfg Config) *Network {
	sd := &SyncerDelegate{}
	var mlCfg *memberlist.Config = cfg
	mlCfg.Delegate = sd
	ml, err := memberlist.Create(mlCfg)
	if err != nil {
		panic(err.Error())
	}

	n := &Network{
		Memberlist:  ml,
		Subscribers: make(map[string][]Callback),
	}
	sd.Callback = n.receiveMessage
	return n
}

func (n *Network) Name() string {
	return n.Memberlist.LocalNode().Name
}

// Join tries to connect to the given peers and make them part of the network.
func (n *Network) Join(peers []string) {

	if len(peers) > 0 {
		n.Memberlist.Join(peers)
	}
}

func (n *Network) Leave(timeout time.Duration) error {
	n.Memberlist.Leave(timeout)

	return n.Memberlist.Shutdown()
}

func (n *Network) Send(peer, messageType, message string) error {
	if printMessages {
		log.Printf("SENDING %s, %s:\t%s\n", peer, messageType, message)
	}

	var peerNode *memberlist.Node
	for _, member := range n.Memberlist.Members() {
		if member.Name == peer {
			peerNode = member
			break
		}
	}

	self := n.Memberlist.LocalNode()
	return n.Memberlist.SendToTCP(peerNode, n.serializeMessage(self.Name, messageType, message))
}

// BroadcastTCP sends a message to each peer in the network, aborting on the first error.
// This message can reach, 0, all or any number of members.
func (n *Network) BroadcastTCP(messageType, message string) error {
	if printMessages {
		log.Printf("BROADCAST %s:\t%s\n", messageType, message)
	}

	self := n.Memberlist.LocalNode()
	data := n.serializeMessage(self.Name, messageType, message)
	for _, member := range n.Memberlist.Members() {
		if member == self {
			continue
		}

		if err := n.Memberlist.SendToUDP(member, data); err != nil {
			return err
		}
	}
	return nil
}

func (n *Network) Subscribe(messageType string, callback func(peer, messageType, message string)) {
	l, ok := n.Subscribers[messageType]
	if !ok {
		l = []Callback{}
	}

	n.Subscribers[messageType] = append(l, callback)
}

func (n *Network) receiveMessage(data []byte) {
	senderPeer, messageType, message := n.deserializeMessage(data)

	if printMessages {
		log.Printf("RECEIVED %s, %s:\t%s\n", senderPeer, messageType, message)
	}

	for _, cb := range n.Subscribers[messageType] {
		go cb(senderPeer, messageType, message)
	}
}

func (n *Network) serializeMessage(sourcePeer, messageType, message string) []byte {
	return []byte(sourcePeer + " " + messageType + " " + message)
}
func (n *Network) deserializeMessage(data []byte) (string, string, string) {
	v := strings.SplitN(string(data), " ", 3)
	return v[0], v[1], v[2]
}

type SyncerDelegate struct {
	Callback func(data []byte)
}

// Memberlist Delete Handlers
// NodeMeta is used to retrieve meta-data about the current node
// when broadcasting an alive message. It's length is limited to
// the given byte size. This metadata is available in the Node structure.
func (sd *SyncerDelegate) NodeMeta(limit int) []byte {
	return []byte{}
}

// NotifyMsg is called when a user-data message is received.
// Care should be taken that this method does not block, since doing
// so would block the entire UDP packet receive loop. Additionally, the byte
// slice may be modified after the call returns, so it should be copied if needed.
func (sd *SyncerDelegate) NotifyMsg(msg []byte) {
	sd.Callback(msg)
}

// GetBroadcasts is called when user data messages can be broadcast.
// It can return a list of buffers to send. Each buffer should assume an
// overhead as provided with a limit on the total byte size allowed.
// The total byte size of the resulting data to send must not exceed
// the limit.
func (sd *SyncerDelegate) GetBroadcasts(overhead, limit int) [][]byte {
	return [][]byte{}
}

// LocalState is used for a TCP Push/Pull. This is sent to
// the remote side in addition to the membership information. Any
// data can be sent here. See MergeRemoteState as well. The `join`
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) LocalState(join bool) []byte {
	return nil
}

// MergeRemoteState is invoked after a TCP Push/Pull. This is the
// state received from the remote side and is the result of the
// remote side's LocalState call. The 'join'
// boolean indicates this is for a join instead of a push/pull.
func (sd *SyncerDelegate) MergeRemoteState(buf []byte, join bool) {

}
