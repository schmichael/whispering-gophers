package main

import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
	"github.com/schmichael/whispering-gophers/util"
)


type DataStruct struct {
	Sender string `json:"sender"`
	Msg    string `json:"message"`
}

type RedisMessage struct {
	Version int `json:"version"`
	Type    string `json:"type"`
	Data    DataStruct `json:"data"`
}

type IncomingStruct struct {
	To string `json:"to"`
	Message    string `json:"message"`
}

type IncomingMessage struct {
	Version int `json:"version"`
	Type    string `json:"type"`
	Data    IncomingStruct `json:"data"`
}

func SendToRedis(message Message) {
	d := DataStruct{
		Sender: message.Addr,
		Msg:    message.Body,
	}
	m := RedisMessage{
		Version: 1,
		Type:    "privmsg",
		Data:    d,
	}

	jBytes, err := json.Marshal(m)
	if err != nil {
		log.Fatal(err)
	}

	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	c.Do("PUBLISH", "in", string(jBytes))
}

func StartRedis() {
    fmt.Println("Starting redis sub")
	c, err := redis.Dial("tcp", ":6379")
    if err != nil {
		log.Fatal(err)
    }
    psc := redis.PubSubConn{c}
	psc.Subscribe("out")
	for {
	    switch v := psc.Receive().(type) {
	    case redis.Message:
	        fmt.Printf("%s: message: %s\n", v.Channel, v.Data)
	        im := &IncomingMessage{};
	        err := json.Unmarshal(v.Data, im)
		    if err != nil {
				log.Print(err)
		    }
	        fmt.Println("Data: ", im.Data)
			m := Message{
				ID:   util.RandomID(),
				Addr: self,
				Body: im.Data.Message,
				Nick: *nick,
			}
			Seen(m.ID)
			broadcast(m)
	    case error:
			log.Fatal(v)
		}
	}
}