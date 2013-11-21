// Whispering Gophers is a simple whispernet written in Go and based off of
// Google's excellent code lab: https://code.google.com/p/whispering-gophers/
package main

import (
	"flag"
	"log"

	"github.com/schmichael/whispering-gophers/util"
	"github.com/schmichael/whispering-gophers/clients"
)

var (
	peerAddr = flag.String("peer", "", "peer host:port")
	bindPort = flag.Int("port", 55555, "port to bind to")
	selfNick = flag.String("nick", "Anonymous Coward", "nickname")
)

func main() {
	flag.Parse()

	clients.PeerAddr = *peerAddr
	clients.BindPort = *bindPort
	clients.SelfNick = *selfNick

	go clients.StartClient()

	l, err := util.ListenWithPort(*bindPort)
	if err != nil {
		log.Fatal(err)
	}
	clients.Self = l.Addr().String()
	log.Println("Listening on", clients.Self)

	go clients.DiscoveryListen()
	go clients.DiscoveryClient()
	go clients.Dial(*peerAddr)
	go clients.ReadInput()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go clients.Serve(c)
	}
}