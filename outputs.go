package main

import (
	"errors"
	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
	"strings"
	"time"
)

type AudioSenderStream interface {
	 Start() error
	 Stop() error
}

func setupAudioOutput (callback func([][]float32), arguments inputArgs) (AudioSenderStream, error) {
	switch arguments.audioOutput {
		case "jack":
			return setupJackDevice(callback)
		case "file":
			return setupFileOutput(callback, arguments.audioOutputFilePath)
	}
	return nil, errors.New("unsupported output type")
}

func setupFileOutput (callback func([][]float32), filepath string) (AudioSenderStream, error) {
	pathSplit := strings.SplitAfter(filepath,".")
	fileFormat := pathSplit[len(pathSplit) - 1]

	switch fileFormat {
		case "opus":
			return setupOpusOutput(filepath)
		default:
			return setupOpusOutput("output.opus")

	}
}

func setupOpusOutput (filepath string) (AudioSenderStream, error) {
	encoder, err := opus.NewEncoder(48e3,2, opus.AppAudio)
	return nil, nil
}

func setupJackDevice(callback func([][]float32)) (stream *portaudio.Stream, err error) {
	devices, err := portaudio.Devices()
	var outputDevice *portaudio.DeviceInfo

	for _, d := range devices {
		println(d.Name)
		if d.Name == "jack" {
			outputDevice = d
		}
	}
	streamDevParams := portaudio.StreamDeviceParameters{
		Device:   outputDevice,
		Channels: 2,
		Latency:  time.Millisecond * 128,
	}
	streamParams := portaudio.StreamParameters{
		Output:          streamDevParams,
		SampleRate:      48e3,
		FramesPerBuffer: 2,
	}
	if err == nil {
		stream, err = portaudio.OpenStream(streamParams, callback)
	}

	return stream, err
}

