// The interfaces of the protocol how peers talk to each other

package libsyncer

import (
	"fmt"
	"time"
)

const (
	MessageAuctionStart = "auction.start"
	MessageAuctionBid   = "auction.bid"
	MessageAuctionEnd   = "auction.end"
)

type Message struct {
	Type   string
	Format string
}

func (m *Message) Serialize(args ...interface{}) string {
	return fmt.Sprintf(m.Format, args...)
}
func (m *Message) Deserialize(msg string, args ...interface{}) {
	fmt.Sscanf(msg, m.Format, args...)
}

var (
	AuctionStartSerializer = &Message{MessageAuctionStart, "%s\t%s\t%s\t%d\t%s"}
	AuctionBidSerializer   = &Message{MessageAuctionBid, "%s\t%g\t%s"}
	AuctionEndSerializer   = &Message{MessageAuctionEnd, "%s\t%s"}
)

type Price float32
type PeerID string
type AuctionID string
type ByteSize uint64

type FileID struct {
	// The volume where the file is located
	VolumeID string

	// The full path inside the volume where the file is located.
	Path string
}

func (file FileID) String() string {
	return fmt.Sprintf("uri:mediasyncer:%s:%s", file.VolumeID, file.Path)
}

func (file FileID) Equals(other FileID) bool {
	return file.VolumeID == other.VolumeID && file.Path == other.Path
}

type FileStats struct {
	Size    ByteSize
	ModTime *time.Time
}

type Transport interface {
	// Peer name of the local node
	Name() string
	Subscribe(messageType string, callback func(peer, messageType, message string))
	BroadcastTCP(messageType, message string) error
	Send(peer, messageType, message string) error
}

type NetworkProtocol struct {
	T Transport
}

func (np *NetworkProtocol) Name() string {
	return np.T.Name()
}

// AuctionStart
func (np *NetworkProtocol) AuctionStart(auctionID AuctionID, file FileID, stats FileStats) error {
	msg := AuctionStartSerializer.Serialize(
		string(auctionID),
		file.VolumeID, file.Path,
		stats.Size, stats.ModTime.Format(time.RFC3339),
	)

	return np.T.BroadcastTCP(MessageAuctionStart, msg)
}

func (np *NetworkProtocol) OnAuctionStart(cb func(peer string, auctionID AuctionID, file FileID, stats FileStats)) {
	np.T.Subscribe(MessageAuctionStart, func(peer, mtype, message string) {
		var auctionID AuctionID
		var file FileID
		var stats FileStats
		var modTime string

		AuctionStartSerializer.Deserialize(message,
			&auctionID,
			&file.VolumeID,
			&file.Path,
			&stats.Size,
			&modTime,
		)
		t, err := time.Parse(time.RFC3339, modTime)
		if err != nil {
			panic("Malformed timestamp from " + peer + ": " + err.Error())
		}
		stats.ModTime = &t
		cb(peer, auctionID, file, stats)
	})
}

func (np *NetworkProtocol) AuctionBid(peer string, auctionID AuctionID, price Price, url string) error {
	return np.T.Send(peer, MessageAuctionBid, AuctionBidSerializer.Serialize(auctionID, float32(price), url))
}

func (np *NetworkProtocol) OnAuctionBid(cb func(peer string, auctionID AuctionID, price Price, url string)) {
	np.T.Subscribe(MessageAuctionBid, func(peer, mtype, msg string) {
		var auctionID AuctionID
		var price Price
		var url string

		AuctionBidSerializer.Deserialize(msg, &auctionID, &price, &url)

		cb(peer, auctionID, price, url)
	})
}

func (np *NetworkProtocol) AuctionEnd(auctionID AuctionID, winnerPeer string) error {
	return np.T.BroadcastTCP(MessageAuctionEnd, AuctionEndSerializer.Serialize(auctionID, winnerPeer))
}

func (np *NetworkProtocol) OnAuctionEnd(cb func(peer string, auctionID AuctionID, winnerPeer string)) {
	np.T.Subscribe(MessageAuctionEnd, func(peer, mtype, msg string) {
		var auctionID AuctionID
		var winnerPeer string
		AuctionEndSerializer.Deserialize(msg, &auctionID, &winnerPeer)
		cb(peer, auctionID, winnerPeer)
	})
}
