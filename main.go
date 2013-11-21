// Whispering Gophers is a simple whispernet written in Go and based off of
// Google's excellent code lab: https://code.google.com/p/whispering-gophers/
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"github.com/schmichael/whispering-gophers/util"
)

var (
	peerAddr = flag.String("peer", "", "peer host:port")
	bindPort = flag.Int("port", 55555, "port to bind to")
	selfNick = flag.String("nick", "", "nickname")
	self     string
	discPort = 5555
)

// Defines a single message sent from one peer to another
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
		log.Fatalf("Unable to listen on port %d: %v", *bindPort, err)
	}
	self = l.Addr().String()
	if *selfNick == "" {
		log.Println("Listening on", self)
	} else {
		log.Printf("Listening on %s as nick %s", self, *selfNick)
	}

	go discoveryListen()
	go discoveryClient()

	if *peerAddr != "" {
		go dial(*peerAddr)
	} else {
		log.Println("No -peer specified. Waiting to receive discovery packets.")
	}
	go readInput()

	for {
		c, err := l.Accept()
		if err != nil {
			log.Fatalf("Unable to accept connection: %v", err)
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
			if err == io.EOF {
				log.Printf("%s disconnected.", c.RemoteAddr())
			} else {
				log.Println(err)
			}
			return
		}
		if Seen(m.ID) {
			continue
		}
		nick := m.Nick
		if nick == "" {
			nick = m.Addr
		}
		fmt.Printf("%s: %s\n", nick, m.Body)
		broadcast(m)
		go dial(m.Addr)
	}
}

func readInput() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		body := s.Text()
		if body != "" {
			m := Message{
				ID:   util.RandomID(),
				Addr: self,
				Body: s.Text(),
				Nick: *selfNick,
			}
			Seen(m.ID)
			broadcast(m)
		}
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
		log.Fatalf("Couldn't send UDP?!?! %v", err)
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
		if e2, ok := err.(*net.OpError); ok && e2.Err.Error() == "address already in use" {
			log.Printf("UDP discovery port %d already in use. Inbound discovery disabled.", discPort)
			return
		} else {
			log.Fatalf("Couldn't open UDP?!? %v", err)
		}
	}
	for {

		data := make([]byte, 0)
		_, _, err := socket.ReadFromUDP(data)
		if err != nil {
			log.Fatal("Problem reading UDP packet: %v", err)
		}
		bcastAddr := string(data)
		if bcastAddr != "" && bcastAddr != self {
			log.Printf("Adding this address to Peer List: %v", bcastAddr)
			dial(bcastAddr)
		}

	}
}
