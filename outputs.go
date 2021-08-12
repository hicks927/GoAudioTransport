package main

import (
	"bytes"
	"errors"
	"github.com/gordonklaus/portaudio"
	"gopkg.in/hraban/opus.v2"
	"os"
	"strings"
	"time"
)

type AudioSenderStream interface {
	Start() error
	Stop() error
	Close() error
}

type OpusFileWriter struct {
	encoder              *opus.Encoder
	file                 *os.File
	inputPcm, opusBuffer *bytes.Buffer
	completeWrite        chan bool
	callback             func([][]float32)
}

func (o OpusFileWriter) Close() error {
	panic("implement me")
}

func (o OpusFileWriter) Start() error {
	panic("implement me")
}

func (o OpusFileWriter) Stop() error {
	panic("implement me")
}

func setupAudioOutput(callback func([][]float32), arguments inputArgs) (AudioSenderStream, error) {
	switch arguments.audioOutput {
	case "jack":
		return setupJackDevice(callback)
	case "file":
		return setupFileOutput(callback, arguments.audioOutputFilePath)
	}
	return nil, errors.New("unsupported output type")
}

func setupFileOutput(callback func([][]float32), filepath string) (AudioSenderStream, error) {
	pathSplit := strings.SplitAfter(filepath, ".")
	fileFormat := pathSplit[len(pathSplit)-1]

	switch fileFormat {
	case "opus":
		return setupOpusOutput(filepath, callback)
	default:
		return setupOpusOutput("output.opus", callback)

	}
}

func setupOpusOutput(filepath string, callback func(out [][]float32)) (OpusFileWriter, error) {
	encoder, err := opus.NewEncoder(48e3, 2, opus.AppAudio)
	file, err := os.OpenFile(filepath, os.O_RDWR, os.ModeExclusive)
	inputPCM := new(bytes.Buffer)
	opusBuffer := new(bytes.Buffer)

	completeWrite := make(chan bool)

	obj := OpusFileWriter{
		encoder:       encoder,
		file:          file,
		inputPcm:      inputPCM,
		opusBuffer:    opusBuffer,
		completeWrite: completeWrite,
		callback:      callback,
	}

	return obj, err
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
