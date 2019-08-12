package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

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
		log.Println("Ready, send me signal to start")
		_, add, _ := conn.ReadFrom(b[:])
		log.Println("starting...")
		start <- struct{}{}
		log.Println("Start to stream audio,have ", ab.q.ReadAvailble(), " samples")

		var i int
		sleep := 10
		if v := os.Getenv("SLEEP"); v != "" {
			sleep, _ = strconv.Atoi(v)
		}
		log.Printf("will sleep %d between writes\n", sleep)
		for {
			v, _ := ab.q.Pop()
			conv.Int16ToBytes(v, b[:], (i*2)%1024)
			i++
			if (i*2)%1024 == 0 {
				conn.WriteTo(b[:], add)
				time.Sleep(time.Duration(sleep) * time.Millisecond)
			}
		}

	}
}

func GetHttpHandler(m *MidiContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snote, ok := r.URL.Query()["note"]

		if !ok || len(snote) < 1 {
			return
		}

		note, err := strconv.Atoi(snote[0])
		if err != nil {
			return
		}

		log.Println("Set note ", note)
		m.notes <- note
	}
}
