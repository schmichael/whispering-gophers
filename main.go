// Whispering Gophers is a simple whispernet written in Go and based off of
// Google's excellent code lab: https://code.google.com/p/whispering-gophers/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"github.com/schmichael/whispering-gophers/util"
)

var (
	peerAddr = flag.String("peer", "", "peer host:port")
	bindPort = flag.Int("port", 55555, "port to bind to")
	nick     = flag.String("nick", "Anonymous Coward", "nickname")
	self     string
	discPort int = 5555
)

type Message struct {
	// Random ID for each message used to prevent re-broadcasting messages
	ID string
	// IP:Port combination the peer who sent a message is listening on
	Addr string
	// Actual message to display
	Body string
	// Nickname
	Nick string `json:"omitempty"`
}

func main() {
	flag.Parse()

	l, err := util.ListenWithPort(*bindPort)
	if err != nil {
		log.Fatal(err)
	}
	self = l.Addr().String()
	log.Println("Listening on", self)

	go discoveryListen()
	go discoveryClient()
	go dial(*peerAddr)
	go readInput()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serve(c)
	}
}

var peers = &Peers{m: make(map[string]chan<- Message)}

type Peers struct {
	m  map[string]chan<- Message
	mu sync.RWMutex
}

// Add creates and returns a new channel for the given peer address.
// If an address already exists in the registry, it returns nil.
func (p *Peers) Add(addr string) <-chan Message {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, ok := p.m[addr]; ok {
		return nil
	}
	ch := make(chan Message)
	p.m[addr] = ch
	return ch
}

// Remove deletes the specified peer from the registry.
func (p *Peers) Remove(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.m, addr)
}

// List returns a slice of all active peer channels.
func (p *Peers) List() []chan<- Message {
	p.mu.RLock()
	defer p.mu.RUnlock()
	l := make([]chan<- Message, 0, len(p.m))
	for _, ch := range p.m {
		l = append(l, ch)
	}
	return l
}

func broadcast(m Message) {
	for _, ch := range peers.List() {
		select {
		case ch <- m:
		default:
			// Okay to drop messages sometimes.
		}
	}
}

func serve(c net.Conn) {
	defer c.Close()
	d := json.NewDecoder(c)
	for {
		var m Message
		err := d.Decode(&m)
		if err != nil {
			log.Println(err)
			return
		}
		if Seen(m.ID) {
			continue
		}
		fmt.Printf("%#v\n", m)
		broadcast(m)
		go dial(m.Addr)
	}
}

func readInput() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		m := Message{
			ID:   util.RandomID(),
			Addr: self,
			Body: s.Text(),
			Nick: *nick,
		}
		Seen(m.ID)
		broadcast(m)
	}
	if err := s.Err(); err != nil {
		log.Fatal(err)
	}
}

func dial(addr string) {
	if addr == self {
		return // Don't try to dial self.
	}

	ch := peers.Add(addr)
	if ch == nil {
		return // Peer already connected.
	}
	defer peers.Remove(addr)

	c, err := net.Dial("tcp", addr)
	if err != nil {
		log.Println(addr, err)
		return
	}
	defer c.Close()

	e := json.NewEncoder(c)
	for m := range ch {
		err := e.Encode(m)
		if err != nil {
			log.Println(addr, err)
			return
		}
	}
}

var seenIDs = struct {
	m map[string]bool
	sync.Mutex
}{m: make(map[string]bool)}

// Seen returns true if the specified id has been seen before.
// If not, it returns false and marks the given id as "seen".
func Seen(id string) bool {
	seenIDs.Lock()
	ok := seenIDs.m[id]
	seenIDs.m[id] = true
	seenIDs.Unlock()
	return ok
}

func discoveryClient() {
	BROADCAST_IPv4 := net.IPv4(255, 255, 255, 255)
	socket, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   BROADCAST_IPv4,
		Port: discPort,
	})
	if err != nil {
		log.Fatal("Couldn't send UDP?!?! %v", err)
	}
	socket.Write([]byte(self))
	log.Printf("Sent a discovery packet!")

}

func discoveryListen() {
	socket, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: discPort,
	})

	if err != nil {
		log.Fatal("Couldn't open UDP?!? %v", err)
	}
	for {

		data, err := ioutil.ReadAll(socket)
		if err != nil {
			log.Fatal("Problem reading UDP packet: %v", err)
		}
		bcastAddr := string(data)
		if bcastAddr != self {
			log.Printf("Adding this address to Peer List: %v", bcastAddr)
			peers.Add(bcastAddr)
		}

	}
}
