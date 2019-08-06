package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cloudnoize/elport"
	locklessq "github.com/cloudnoize/locklessQ"

	"github.com/google/gousb"
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
	if op == "udp" {
		log.Println("UDP mode...")
		start := make(chan struct{})
		go playMidi(recch, dur, start)
		ServeUdp(addr, ab, start, stream)
	} else if op == "play" {
		go playMidi(recch, dur, nil)
		<-done
		Play(ab, sf, uint64(desiredSR), channels, dur)

	} else if op == "save" {
		go playMidi(recch, dur, nil)
		<-done
		saveWav(ab, uint32(desiredSR), wavPath+GetFileName())
		http.ListenAndServe(addr, nil)
	}

	pa.Terminate()
}

func playMidi(recch chan struct{}, dur int, start chan struct{}) {
	ctx := gousb.NewContext()
	defer ctx.Close()

	// Open any device with a given VID/PID using a convenience function.
	dev, err := ctx.OpenDeviceWithVIDPID(gousb.ID(0xfc02), gousb.ID(0x0101))
	if err != nil {
		log.Fatalf("Could not open a device: %v", err)
	}
	if dev == nil {
		log.Fatalf("dev is null	")
	}
	dev.SetAutoDetach(true)
	defer dev.Close()
	// Switch the configuration to #2.
	cfg, err := dev.Config(1)
	if err != nil {
		log.Fatalf("%s.Config(2): %v", dev, err)
	}
	defer cfg.Close()

	intf, err := cfg.Interface(1, 0)
	if err != nil {
		log.Fatalf("%s.Interface(3, 0): %v", cfg, err)
	}
	defer intf.Close()
	println(intf.String())

	// Open an OUT endpoint.
	ep, err := intf.OutEndpoint(2)
	if err != nil {
		log.Fatalf("%s.OutEndpoint(7): %v", intf, err)
	}

	on := []byte{9, 144, 60, 110}
	off := []byte{9, 144, 60, 0}

	if start != nil {
		log.Println("Waiting for signal")
		<-start
	}

	recch <- struct{}{}
	for i := 0; i < dur; i++ {
		ep.Write(off[:])
		time.Sleep(10 * time.Millisecond)
		writeBytes, err := ep.Write(on[:])
		if err != nil {
			fmt.Println("Write returned an error:", err)
		}
		if writeBytes != len(on) {
			log.Fatalf("data out of %d sent", writeBytes)
		}
		time.Sleep(990 * time.Millisecond)
	}
	println("Finish midi")
	recch <- struct{}{}
	ep.Write(off[:])
}
