package libsyncer

import (
	"log"
	"os"
)

type Bidder struct {
	Volume       Volume
	Network      NetworkProtocol
	PriceFormula PriceFormula
	FileServer   *FileServer

	active   bool
	auctions chan bidderAuctionStarted
}

type bidderAuctionStarted struct {
	peer  string
	ID    AuctionID
	file  FileID
	stats FileStats
}

func NewBidder(n NetworkProtocol, vol Volume, pf PriceFormula, fs *FileServer) *Bidder {
	b := &Bidder{
		Network:      n,
		Volume:       vol,
		PriceFormula: pf,
		FileServer:   fs,

		active:   true,
		auctions: make(chan bidderAuctionStarted),
	}

	b.Network.OnAuctionStart(func(peer string, auctionID AuctionID, file FileID, stats FileStats) {
		b.auctions <- bidderAuctionStarted{peer, auctionID, file, stats}
	})

	return b
}

func (b *Bidder) Serve() {
	for b.active {
		select {
		case auction := <-b.auctions:
			log.Println("Received auction " + string(auction.ID) + " from " + auction.peer + " for file " + auction.file.String())
			freeSpace := ByteSize(b.Volume.AvailableBytes())
			if freeSpace < auction.stats.Size {
				log.Println(auction.ID + ": not bidding - not enough space on volume.")
				continue
			}
			price := b.PriceFormula(auction.file, auction.stats, freeSpace)

			if price == -1 {
				log.Println("Not bidding. File not wanted.")
				continue
			}

			_, err := b.Volume.Stat(auction.file.Path)
			if err != nil {
				if os.IsNotExist(err) {
					// Only bid, if we don't have this file locally.
					url, err := b.FileServer.CreateUploadURL(FileID{
						VolumeID: b.Volume.ID(),
						Path:     auction.file.Path,
					})
					if err != nil {
						panic("Unable to create upload URL")
					}
					b.Network.AuctionBid(auction.peer, auction.ID, price, url)
					continue
				}
				panic("Stat error: " + err.Error())
			}
			log.Println("Ignoring - file exists locally.")
		}
	}
}

func (b *Bidder) Stop() {
	b.active = false
}
