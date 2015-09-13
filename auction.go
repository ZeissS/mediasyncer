package mediasyncer

import (
	"fmt"
	"log"
	"os"
	"time"
)

const AuctionTimeout = 5 * time.Second

type Auctioneer struct {
	Network      NetworkProtocol
	PriceFormula PriceFormula
	Volume       Volume
	Ticker       *time.Ticker
	Uploader     *Uploader

	Bids              chan auctionBid
	UploadsInProgress map[string]struct{}
	UploadDone        chan FileID
}

type auctionBid struct {
	peer      string
	auctionID AuctionID
	price     Price
	uploadURL string
}

func NewAuctioneer(n NetworkProtocol, priceFormula PriceFormula, vol Volume, uploader *Uploader) *Auctioneer {
	a := &Auctioneer{
		Network:      n,
		Ticker:       time.NewTicker(10 * time.Second),
		PriceFormula: priceFormula,
		Uploader:     uploader,
		Volume:       vol,

		Bids:              make(chan auctionBid),
		UploadsInProgress: make(map[string]struct{}),
		UploadDone:        make(chan FileID),
	}

	n.OnAuctionBid(func(peer string, auctionID AuctionID, price Price, url string) {
		a.Bids <- auctionBid{peer, auctionID, price, url}
	})

	return a
}

type auctionCanidate struct {
	file  FileID
	stats FileStats
	price Price
}

func (a *Auctioneer) collectFileList() []auctionCanidate {
	var canidates []auctionCanidate
	freeSpace := ByteSize(a.Volume.AvailableBytes())

	a.Volume.Walk(func(fullpath string, info os.FileInfo, err error) error {
		if info.Size() == 0 {
			return nil
		}

		file := FileID{
			VolumeID: a.Volume.ID(),
			Path:     fullpath,
		}

		if _, ok := a.UploadsInProgress[file.String()]; ok {
			return nil
		}

		t := info.ModTime()

		if t.After(time.Now().Add(-1 * 60 * time.Minute)) {
			//log.Printf("Skipping %s - too young.\n", fullpath)
			return nil
		}
		stats := FileStats{
			Size:    ByteSize(info.Size()),
			ModTime: &t,
		}
		price := a.PriceFormula(file, stats, freeSpace)

		canidates = append(canidates, auctionCanidate{
			file:  file,
			stats: stats,
			price: price,
		})
		return nil
	})

	return canidates
}

func (a *Auctioneer) Serve() {
	auctionSeq := 0

	auctionInProgress := false
	var auctionID AuctionID
	var auctionEndTimer <-chan time.Time
	var bids []auctionBid
	var auctionCanidate auctionCanidate

	for {
		select {
		case <-a.Ticker.C:
			if auctionInProgress {
				log.Println("Ignoring auction tick - auction-in-progress.")
				continue
			}

			canidates := a.collectFileList()
			if len(canidates) == 0 {
				log.Println("Ignoring auction tick - no local file to auction found.")
				continue
			}

			auctionInProgress = true
			auctionID = AuctionID(fmt.Sprintf("%s/auction/%d", a.Network.Name(), auctionSeq))
			auctionSeq++

			auctionCanidate = canidates[0]

			a.Network.AuctionStart(auctionID, auctionCanidate.file, auctionCanidate.stats)
			auctionEndTimer = time.After(AuctionTimeout)

		case bid := <-a.Bids:
			if !auctionInProgress {
				return
			}
			if auctionID != bid.auctionID {
				return
			}

			bids = append(bids, bid)

		case <-auctionEndTimer:
			if len(bids) == 0 {
				log.Println("No bids received. Auction failed.")
			} else {
				var winningBid auctionBid = bids[0]

				for _, bid := range bids {
					if bid.price > winningBid.price {
						winningBid = bid
					}
				}

				log.Printf("# Auction ended. %d bids received.\n", len(bids))
				log.Printf("# File: %v\n", auctionCanidate.file)
				if winningBid.price > auctionCanidate.price {
					log.Printf("# Peer %s won the auction with %v\n", winningBid.peer, winningBid.price)

					a.Network.AuctionEnd(auctionID, winningBid.peer)

					a.UploadsInProgress[auctionCanidate.file.String()] = struct{}{}
					go a.Uploader.Upload(auctionCanidate.file, PeerID(winningBid.peer), winningBid.uploadURL, a.UploadDone)
				} else {
					log.Printf("# Keeping file locally. No remote winner found (highest: %v from %s)\n", winningBid.price, winningBid.peer)
					a.Network.AuctionEnd(auctionID, a.Network.Name())
				}
			}

			auctionInProgress = false
			bids = nil

		case file := <-a.UploadDone:
			log.Printf("# Upload finished: %s\n", file)
			if err := a.Volume.Delete(file.Path); err != nil {
				panic("delete failed: " + err.Error())
			}
			delete(a.UploadsInProgress, file.String())
		}
	}
}

func (a *Auctioneer) Stop() {
	a.Ticker.Stop()
}
