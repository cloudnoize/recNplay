package main

import (
	"bufio"
	"fmt"
	"log"
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
	Recored(ab, sf, uint64(desiredSR), channels, recch, done)
	go playMidi(recch, dur)
	println("Before done")
	<-done
	println("Play")
	Play(ab, sf, uint64(desiredSR), channels, dur)
	pa.Terminate()
}

func playMidi(recch chan struct{}, dur int) {
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
	recch <- struct{}{}
	println("Finish midi")
	ep.Write(off[:])
}
