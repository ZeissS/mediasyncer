package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ogier/pflag"

	"github.com/zeisss/mediasyncer/libsyncer"
	"github.com/zeisss/mediasyncer/disk"
	"github.com/zeisss/mediasyncer/p2p"
)

var p2pConfig p2p.Config = p2p.DefaultConfig()
var fsConfig libsyncer.FileServerConfig

var (
	volumePath           string
	formula              string
	formulaStaticPrice   float32
	printNetworkMessages bool
)

func init() {
	pflag.StringVar(&formula, "price-formula", "static", "What price formular to use? static, random")
	pflag.Float32Var(&formulaStaticPrice, "price-static", 1.0, "Price for static formular")

	pflag.StringVar(&volumePath, "volume", "./lib", "What files to sync")

	pflag.StringVar(&fsConfig.Addr, "http-addr", "127.0.0.1", "IP to listen on. Must be resolvable by all peers")
	pflag.IntVar(&fsConfig.Port, "http-port", 8080, "Port for HTTP FileServer")

	pflag.IntVar(&p2pConfig.BindPort, "bind-port", 8000, "The port to bind to")
	pflag.StringVar(&p2pConfig.Name, "name", "mediasyncer", "The name of this process. Must be unique for the memberlist cluster")

	pflag.BoolVar(&printNetworkMessages, "debug", false, "Print network messages received/sent")
}

func pricer() libsyncer.PriceFormula {
	switch formula {
	case "static":
		return libsyncer.PriceFormulaStatic(libsyncer.Price(formulaStaticPrice))
	case "random":
		return libsyncer.PriceFormulaRandom()
	default:
		panic("Unknown formula: " + formula)
	}
}

func volume() libsyncer.Volume {
	v := disk.Open(volumePath)
	return v
}

func main() {
	pflag.Parse()

	p2p.PrintMessages(printNetworkMessages)

	log.SetPrefix(p2pConfig.Name + " ")

	network := p2p.New(p2pConfig)
	network.Join(pflag.Args())

	cfg := libsyncer.Config{
		FileServerConfig: fsConfig,
		PriceFormula:     pricer(),
		Transport:        network,
		Volume:           volume(),
	}
	syncer := libsyncer.New(cfg)
	go syncer.Serve()

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	log.Println("Received shutdown signal. Stopping ...")
	syncer.Stop()

	if err := network.Leave(10 * time.Second); err != nil {
		log.Printf("ERROR: %v", err)
	}
}
