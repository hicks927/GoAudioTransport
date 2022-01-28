package main

import (
	"encoding/binary"
	"errors"
	"github.com/gordonklaus/portaudio"
	"github.com/hajimehoshi/go-mp3"
	"math"
	"os"
	"time"
)

func getInputCallback(inputType string, arguments inputArgs) (func([][]float32), error) {
	switch inputType {
	case "sine":
		return generateSineCallback(arguments.audioSineFreq, 48e3)
	case "file":
		return generateMp3FileCallback(arguments.audioInputFilePath)
	case "jack":
		rb := newRingBuffer(48e3*0.2, 2) //TODO Make buffer length modifiable
		return generateJackCallback(rb)
	}
	return nil, errors.New("invalid input type")
}

func generateJackCallback(ringBuffer RingBufferF) (func([][]float32), error) {
	devices, err := portaudio.Devices()
	var inputDevice *portaudio.DeviceInfo

	for _, d := range devices {
		println(d.Name)
		if d.Name == "jack" {
			inputDevice = d
		}
	}

	streamDevParams := portaudio.StreamDeviceParameters{
		Device:   inputDevice,
		Channels: 2,
		Latency:  time.Millisecond * 128,
	}

	streamParams := portaudio.StreamParameters{
		Input:           streamDevParams,
		SampleRate:      48e3,
		FramesPerBuffer: 2,
	}

	stream, err := portaudio.OpenStream(streamParams, ringBuffer.writeAudio)
	err = stream.Start()
	time.Sleep(20 * time.Millisecond)

	return func(i [][]float32) {
		ringBuffer.readAudio(i)
	}, err

}

func generateSineCallback(sineFreq float64, sampleRate float64) (func([][]float32), error) {
	var err error
	var xPhase = float64(0)
	var xStep = sineFreq / sampleRate

	return func(out [][]float32) {
		for i := range out[0] {
			var sample float32
			sample = float32(math.Sin(2 * math.Pi * xPhase))
			out[0][i] = sample
			out[1][i] = sample
			_, xPhase = math.Modf(xPhase + xStep)
		}
	}, err
}

func generateMp3FileCallback(filePath string) (func([][]float32), error) {
	var err error
	r, err := os.Open(filePath)
	decoder, err := mp3.NewDecoder(r)
	return func(out [][]float32) {
		//samples := len(out[0])
		for i := range out[0] {
			audio := make([]byte, 4)
			_, err = decoder.Read(audio)
			sampleL := int16(binary.LittleEndian.Uint16(audio[0:2]))
			sampleR := int16(binary.LittleEndian.Uint16(audio[2:4]))
			out[0][i] = float32(sampleL) / float32(math.MaxInt16)
			out[1][i] = float32(sampleR) / float32(math.MaxInt16)
		}
	}, err
}
