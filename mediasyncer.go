package mediasyncer

import (
	"sync"
)

type Config struct {
	Transport        Transport
	PriceFormula     PriceFormula
	Volume           Volume
	FileServerConfig FileServerConfig
}
type Syncer struct {
	Config
	running sync.WaitGroup

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

		Auctioneer: auctioneer,
		FileServer: fs,
		Bidder:     bidder,
	}
}

func (s *Syncer) Serve() {
	go s.FileServer.Serve()
	go s.Auctioneer.Serve()
	go s.Bidder.Serve()

}

func (s *Syncer) Stop() {
	s.Auctioneer.Stop()
	s.Bidder.Stop()
	s.FileServer.Close()

	s.running.Wait()
}
