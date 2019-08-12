package main

import (
	"fmt"
	"log"
	"time"
)

type MidiContext struct {
	note    int
	noteOn  []byte
	noteOff []byte
	notes   chan int
}

func NewMidiContext() *MidiContext {
	m := &MidiContext{note: 60, noteOn: []byte{9, 144, 60, 110}, noteOff: []byte{9, 144, 60, 0}, notes: make(chan int)}
	return m
}

func (m *MidiContext) SetNote() {
	for {
		note := <-m.notes
		m.note = note
	}
}
func (m *MidiContext) playMidi(recch chan struct{}, dur int, start chan struct{}) {
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

	if start != nil {
		log.Println("Waiting for signal")
		<-start
	}

	var note byte

	recch <- struct{}{}
	for i := 0; i < dur; i++ {
		ep.Write(m.noteOff[:])
		time.Sleep(10 * time.Millisecond)
		note = byte(m.note)
		m.noteOn[2] = note
		m.noteOff[2] = note
		writeBytes, err := ep.Write(m.noteOn[:])
		if err != nil {
			fmt.Println("Write returned an error:", err)
		}
		if writeBytes != len(m.noteOn) {
			log.Fatalf("data out of %d sent", writeBytes)
		}
		time.Sleep(990 * time.Millisecond)
	}
	println("Finish midi")
	recch <- struct{}{}
	ep.Write(m.noteOff[:])
}
