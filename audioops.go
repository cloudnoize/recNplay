package main

import (
	"fmt"
	"time"
	"unsafe"

	"github.com/cloudnoize/elport"

	locklessq "github.com/cloudnoize/locklessQ"
)

type AudioBuffer struct {
	q        *locklessq.Qint16
	isRecord bool
}

func (r *AudioBuffer) CallBack(inputBuffer, outputBuffer unsafe.Pointer, frames uint64) {
	if r.isRecord {
		r.RecordCallBack(inputBuffer, outputBuffer, frames)
		return
	}
	r.PlayCallBack(inputBuffer, outputBuffer, frames)
}

func (r *AudioBuffer) RecordCallBack(inputBuffer, outputBuffer unsafe.Pointer, frames uint64) {
	ib := (*[1024]int16)(inputBuffer)
	for i := 0; i < len(ib); i++ {
		r.q.Insert((*ib)[i])
	}
}

func (r *AudioBuffer) PlayCallBack(inputBuffer, outputBuffer unsafe.Pointer, frames uint64) {
	ob := (*[1024]int16)(outputBuffer)
	for i := 0; i < int(frames); i++ {
		(*ob)[i], _ = r.q.Pop()
	}
}

func Recored(ab *AudioBuffer, sf pa.SampleFormat, desiredSR uint64, channels int, recch chan struct{}, done chan struct{}) {
	var (
		in pa.PaStreamParameters
	)

	devnum, err := selectDevice("Select device num for record: ")
	if err != nil {
		fmt.Println(err)
		return
	}

	println("selected  ", devnum, " sample format ", sf)

	in = pa.PaStreamParameters{DeviceNum: devnum, ChannelCount: channels, Sampleformat: sf}

	err = pa.IsformatSupported(&in, nil, desiredSR)

	if err != nil {
		println("ERROR ", err.Error())
		return
	}

	fmt.Println(in, " supports ", desiredSR)

	pa.CbStream = ab

	//Open stream
	s, err := pa.OpenStream(&in, nil, in.Sampleformat, desiredSR, 1024)
	if err != nil {
		println("ERROR ", err.Error())
		return
	}

	go func() {
		<-recch
		println("recording...")

		s.Start()

		<-recch
		println("Stop recording...")
		s.Stop()
		s.Close()
		println("Send done...")
		done <- struct{}{}

	}()

}

func Play(ab *AudioBuffer, sf pa.SampleFormat, desiredSR uint64, channels int, dur int) {

	devnum, err := selectDevice("Select device num for play: ")
	ab.isRecord = false
	if err != nil {
		fmt.Println(err)
		return
	}

	println("selected  ", devnum, " sample format ", sf)

	out := pa.PaStreamParameters{DeviceNum: devnum, ChannelCount: channels, Sampleformat: sf}

	err = pa.IsformatSupported(nil, &out, desiredSR)

	if err != nil {
		println("ERROR ", err.Error())
		return
	}

	fmt.Println(out, " supports ", desiredSR)

	pa.CbStream = ab

	//Open stream
	s, err := pa.OpenStream(nil, &out, out.Sampleformat, desiredSR, 1024)
	if err != nil {
		println("ERROR ", err.Error())
		return
	}

	println("playing...")

	s.Start()

	time.Sleep(time.Duration(dur) * time.Second)
	println("Stop playing...")
	s.Stop()
	s.Close()

}
