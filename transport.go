package main

import (
	"log"
	"net"

	"github.com/cloudnoize/conv"
)

func ServeUdp(addr string, ab *AudioBuffer, start chan struct{}, stream chan struct{}) {
	conn, e := net.ListenPacket("udp", addr)
	if e != nil {
		println(e.Error())
		return
	}

	for {
		var b [1024]byte
		_, add, _ := conn.ReadFrom(b[:])

		log.Println("Ready, send me signal to start")
		conn.ReadFrom(b[:])
		log.Println("starting...")
		start <- struct{}{}
		<-stream
		log.Println("Start to stream audio,have ", ab.q.ReadAvailble(), " samples")
		for {
			var ok bool
			var s int16
			var i int
			for i = 0; i < 512; i++ {
				s, ok = ab.q.Pop()
				if !ok {
					break
				}
				conv.Int16ToBigBytes(s, b[:], i*2)
			}
			//log.Printf("Writing %d samples\n", i)
			conn.WriteTo(b[:i], add)
		}

	}
}
