package mediasyncer

import (
	"sync"
	"time"
)

type Config struct {
	Transport        Transport
	PriceFormula     PriceFormula
	Volume           *Volume
	FileServerConfig FileServerConfig
}
type Syncer struct {
	Config
	running sync.WaitGroup
	ticker  *time.Ticker

	FileServer *FileServer
	Bidder     *Bidder
	Auctioneer *Auctioneer
}

func New(cfg Config) *Syncer {
	proto := NetworkProtocol{cfg.Transport}

	fs := NewFileServer(cfg.FileServerConfig, cfg.Volume)

	uploader := &Uploader{cfg.Volume}
	auctioneer := NewAuctioneer(proto, cfg.PriceFormula, cfg.Volume, uploader)
	bidder := NewBidder(proto, cfg.Volume, cfg.PriceFormula, fs)

	return &Syncer{
		Config: cfg,
		ticker: time.NewTicker(10 * time.Second),

		Auctioneer: auctioneer,
		FileServer: fs,
		Bidder:     bidder,
	}
}

func (s *Syncer) Serve() {
	go s.FileServer.Serve()
	go s.Auctioneer.Serve()
	go s.Bidder.Serve()

	/*s.Transport.Subscribe("hello", func(peer, mtype, msg string) {
		log.Println("!hello", peer, mtype, msg)
	})

	for range s.ticker.C {
		s.Tick()
	}*/
}

func (s *Syncer) Tick() {
	s.running.Add(1)
	defer s.running.Done()
	s.Transport.BroadcastTCP("hello", "Hello from "+s.Transport.Name())
}

func (s *Syncer) Stop() {

	s.ticker.Stop()
	s.Auctioneer.Stop()
	s.Bidder.Stop()
	s.FileServer.Close()

	s.running.Wait()
}
