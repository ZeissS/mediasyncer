package libsyncer

import (
	"log"
	"os"
)

// The Bidder is a service that subscribes to AuctionStarted events on the NetworkProtocol,
// calculates a bid with the PriceFormula and responds with a Bid.
// If not enough space is available on the Volume, the auction is ignored.
// If the PriceFormula returns a negative price, the auction is ignored.
type Bidder struct {
	volume       Volume
	network      NetworkProtocol
	priceFormula PriceFormula
	fileServer   *FileServer

	active   bool
	auctions chan bidderAuctionStarted
}

// bidderAuctionStarted represents an internal message which is generated for
// new auction events.
type bidderAuctionStarted struct {
	peer  string
	ID    AuctionID
	file  FileID
	stats FileStats
}

// NewBidder creates a new Bidder for the given dependencies. The bidder is not started yet,
// but immediately subscribes to the NetworkProtocols OnAuctionStart.
func NewBidder(n NetworkProtocol, vol Volume, pf PriceFormula, fs *FileServer) *Bidder {
	b := &Bidder{
		network:      n,
		volume:       vol,
		priceFormula: pf,
		fileServer:   fs,

		active:   true,
		auctions: make(chan bidderAuctionStarted),
	}

	b.network.OnAuctionStart(func(peer string, auctionID AuctionID, file FileID, stats FileStats) {
		b.auctions <- bidderAuctionStarted{peer, auctionID, file, stats}
	})

	return b
}

// Serve runs the bidder loop by consuming an internal chan to react to new auctions.
// Each auction is handled sequentialy.
func (b *Bidder) Serve() {
	for b.active {
		select {
		case auction := <-b.auctions:
			log.Println("Received auction " + string(auction.ID) + " from " + auction.peer + " for file " + auction.file.String())
			freeSpace := ByteSize(b.volume.AvailableBytes())
			if freeSpace < auction.stats.Size {
				log.Println(auction.ID + ": not bidding - not enough space on volume.")
				continue
			}
			price := b.priceFormula(auction.file, auction.stats, freeSpace)

			if price == -1 {
				log.Println("Not bidding. File not wanted.")
				continue
			}

			_, err := b.volume.Stat(auction.file.Path)
			if err != nil {
				if os.IsNotExist(err) {
					// Only bid, if we don't have this file locally.
					url, err := b.fileServer.CreateUploadURL(FileID{
						VolumeID: b.volume.ID(),
						Path:     auction.file.Path,
					})
					if err != nil {
						panic("Unable to create upload URL")
					}
					b.network.AuctionBid(auction.peer, auction.ID, price, url)
					continue
				}
				panic("Stat error: " + err.Error())
			}
			log.Println("Ignoring - file exists locally.")
		}
	}
}

// Stop sets an internal flag to stop the bidder loop in Bidder.Serve()
func (b *Bidder) Stop() {
	b.active = false
}
