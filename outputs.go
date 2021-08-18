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
	opusBuffer, inputPcm *bytes.Buffer

	completeWrite chan bool
	persistData   chan int
	killWriter    chan bool

	callback func([][]float32)
}

func (o OpusFileWriter) Close() error {
	o.completeWrite <- true
	return nil
}

func (o OpusFileWriter) Start() error {
	o.persistData = make(chan int)
	o.killWriter = make(chan bool)
	go func() {
		for {
			select {
			case <-o.killWriter:
				o.file.Close()
				return
			case byteNo := <-o.persistData:
				o.file.Write(o.opusBuffer.Next(byteNo))
			}
		}
	}()

	go func() {
		a := make([][]float32, 2)
		for i := range a {
			a[i] = make([]float32, 512) //make this a switchable buffer thing
		}
		poll := time.Tick(time.Second * 512 / 48e3)
		for {
			select {
			case <-o.completeWrite:
				encodedBytes := make([]byte, 960*4)
				requiredPadding := (960 * 4) - o.inputPcm.Len()
				padding := make([]byte, requiredPadding)
				o.inputPcm.Write(padding)
				o.encoder.EncodeFloat32(bufferToFloat32Arr(o.inputPcm, 960), encodedBytes)
				o.persistData <- 960 * 4

				o.killWriter <- true
				return

			case <-poll:
				o.callback(a)
				o.inputPcm.Write(float32Arr2DToByteArr(a))

				if o.inputPcm.Len() > 48e3/100*2*4 { //Hardcoded for 48K, 2 channels and 10ms
					encodedBytes := make([]byte, 960*4)
					_, err := o.encoder.EncodeFloat32(bufferToFloat32Arr(o.inputPcm, 960), encodedBytes)

					if err != nil {
						panic(err)
					}

					o.opusBuffer.Write(encodedBytes)
					o.persistData <- 960 * 4
				}
			}
		}
	}()

	return nil
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
	//file, err := os.OpenFile(filepath, os.O_WRONLY, os.ModeExclusive)
	file, err := os.Create(filepath)
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
