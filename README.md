Whispering Gophers
==================

A whispernet written in Go.

Based on: https://code.google.com/p/whispering-gophers/

Protocol
--------

TODO

Extensions
----------

Roughly ordered based on difficulty.

1. Drop old messages - requires *Timestamp* field
1. Forget old messages (the basic daemon slowly uses all memory) - may use *Timestamp* field - see below
1. Base ID on hash of Addr + Body + Timestamp - see below
1. Discover based on [UDP broadcasts](https://groups.google.com/d/msg/golang-nuts/nbmYWwHCgPc/ZBw2uH6Bdi4J) See below
1. Build web interface into daemon
 1. Simple input form (textbox + submit button)
 1. Chat output (refreshed on a timer or using [websockets](http://godoc.org/code.google.com/p/go.net/websocket) if you're really fancy)
 1. Peer connection controls (connect to a new host, disconnect from a host)
1. File transfers - see below
1. Send backlog to newly connected peers - see backlog below
1. Flood control: Disconnect and blacklist peers whose activity exceeds a threshold
1. PKI all the things
1. Scalable routing

**UDP Broadcast Protocol**

Broadcasts should be of the form: "<IP>:<PORT>" and may be sent on behalf of any peers.

**Timestamp field**

To support various features, messages should include an optional Timestamp field of the format milliseconds since UNIX epoch as a number. For example:

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
