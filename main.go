package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cloudnoize/elport"
	locklessq "github.com/cloudnoize/locklessQ"
)

func selectDevice(str string) (int, error) {
	pa.ListDevices()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print(str)
	text, _ := reader.ReadString('\n')
	devnum, err := strconv.Atoi(strings.TrimSuffix(text, "\n"))
	if err != nil {
		return 0, nil
	}
	return devnum, nil
}

func main() {
	dur := 10
	if v := os.Getenv("DURATION"); v != "" {
		dur, _ = strconv.Atoi(v)
	}

	op := "play"
	if v := os.Getenv("OP"); v != "" {
		op = v
	}
	log.Println(op)

	addr := ":8765"
	if v := os.Getenv("ADDR"); v != "" {
		addr = v
	}

	wavPath := "/home/eranl/WAV_FILES/server/"

	http.Handle("/", http.FileServer(http.Dir(wavPath)))

	log.SetFlags(log.LstdFlags | log.Llongfile)
	//16 bit
	sf := pa.SampleFormat(8)

	err := pa.Initialize()
	if err != nil {
		println("ERROR ", err.Error())
		return
	}
	desiredSR := 48000
	channels := 1
	ab := &AudioBuffer{q: locklessq.NewQint16(int32(desiredSR * (dur))), isRecord: true}

	done := make(chan struct{})
	recch := make(chan struct{})
	stream := make(chan struct{}, 1)
	Recored(ab, sf, uint64(desiredSR), channels, recch, done, stream)
	midi := NewMidiContext()
	if op == "udp" {
		log.Println("UDP mode...")
		start := make(chan struct{})
		go midi.playMidi(recch, dur, start)
		go http.ListenAndServe(addr, GetHttpHandler(midi))
		ServeUdp(addr, ab, start, stream)
	} else if op == "play" {
		go midi.playMidi(recch, dur, nil)
		<-done
		Play(ab, sf, uint64(desiredSR), channels, dur)

	} else if op == "save" {
		go midi.playMidi(recch, dur, nil)
		<-done
		saveWav(ab, uint32(desiredSR), wavPath+GetFileName())
		http.ListenAndServe(addr, nil)
	}

	pa.Terminate()
}
