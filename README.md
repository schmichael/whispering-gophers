Whispering Gophers
==================

A whispernet written in Go.

Based on: https://code.google.com/p/whispering-gophers/

Getting Started
---------------
If you have not worked in GO before (i.e., don't have a GOPATH, then start here:

  mkdir ~/somedir
  cd somedir
  export GOPATH=~/somedir
  git clone https://github.com/pdxgo/whispering-gophers.git
  go get github.com/pdxgo/whispering-gophers
  cd whispering-gophers
  go build


Running
-------

For the first peer:

```sh
go run main.go -nick phil
```

Next peers:

```sh
go run main.go -peer $IP_OF_PEER:55555 -nick goldy
```

Where ``$IP_OF_PEER`` is the IP address printed by a running peer.

Protocol
--------

*Peers* are long running programs which accept user input and display messages
from other peers. At the minimum a well participating peer must:

* Manually connect to at least 1 peer
* Bind to an address and accept connections from any peer

*Messages* should be of the form:

```json
{
    "ID": "<globally unique string>",
    "Addr": "<IP:Port the sender is listening on>",
    "Body": "<the actual message to display>"
}
```

*Peers* should follow the following behavior with regard to *messages*:

* Display messages (``Body`` key) from other peers
* Disconnect peers which send malformed messages
* Accept user input as ``Body`` content
* Send messages to all known peers.
* When sending messages ``Addr`` must be the ``IP:Port`` combination the peer
  is listening on.
* Connect to new peers seen in ``Addr`` fields of received messages
* Broadcast all messages to all peers (except ``Addr``); store list of Seen
  messages by ID to avoid rebroadcasting
* Payloads without a ``Body`` and/or with unknown keys should be silently
  *ignored* to support extensions
* *Messages are* **not** *delimited or framed.* Multiple messages *may* be sent
  across a single connection. Peers recieving messages *must* detect the end of
  the JSON object to know when one message ends and another may begin.

Connections are unidirectional. In a healthy whispernet each peer will have 1
outgoing connection to send messages to every other peer, as well as 1 incoming
connection from each peer for receiving messages.

Extensions
----------

Roughly ordered based on difficulty.

1. ~~Pretty display messages~~ Done!
1. ~~Nicks!~~ Done!
1. Forget old message IDs (the basic daemon slowly uses all memory) - may use
   *Timestamp* field - see below
1. Base ID on hash of Addr + Body + Timestamp - see below
1. Discover based on [UDP
   broadcasts](https://groups.google.com/d/msg/golang-nuts/nbmYWwHCgPc/ZBw2uH6Bdi4J)
See below - *Partially done*
1. Build web interface into daemon
 1. Simple input form (textbox + submit button)
 1. Chat output (refreshed on a timer or using
    [websockets](http://godoc.org/code.google.com/p/go.net/websocket) if you're
    really fancy)
 1. Peer connection controls (connect to a new host, disconnect from a host)
1. File transfers - see below
1. Send backlog to newly connected peers - see backlog below
1. Flood control: Disconnect and blacklist peers whose activity exceeds a
   threshold
1. PKI all the things
1. Scalable routing

**UDP Broadcast Protocol**

Broadcasts should be of the form: "IP:PORT" (same as ``Addr``) and may be sent on 
behalf of any peers.

**Timestamp field**

To support various features, messages should include an optional Timestamp
field of the format milliseconds since UNIX epoch as a number. For example:

```json
{
    "ID": "abc123",
    "Addr": "192.168.1.53:2345",
    "Body": "Hello world!",
    "Timestamp": 1382050782085
}
```

**Hashed ID field**

To support very rudimentary payload verification, use a hash of the concatenated Addr, Body, and Timestamp fields. To allow backward compatibility with old clients, hashed IDs should be of the form:

```json
{
    "ID": "hash:sha256:<hash>",
    ...
}
```

Where each string starts with "hash:" followed by the name of the hash function used, followed by a colon and the hex encoded hash.

Unknown hash functions should cause the reader to revert to the basic random-string-ID behavior.

Peers should drop messages whose hashes don't match the current message payload.

**File transfers**

Two new keys: "FileType" and "FileContents"

* *FileType* should be the MIME type like: "image/gif"
* *FileContents* should be the Base64 encoded file contents

**Backlog**

Upon receiving a new connection from a peer, a daemon should replay that peer the last N messages (the actual number shouldn't be important) to the peer.
