package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

func float32ToByte(f float32) []byte {
	var buf bytes.Buffer
	err := binary.Write(&buf, binary.LittleEndian, f)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	return buf.Bytes()
}

func float32Arr2DToByteArr(f [][]float32) []byte {
	var buf bytes.Buffer
	for sample := range f[0] {
		buf.Write(float32ToByte(f[0][sample]))
		buf.Write(float32ToByte(f[1][sample]))
	}
	return buf.Bytes()
}

func bufferToFloat32Arr(buf *bytes.Buffer, sliceLength int) []float32 {
	out := make([]float32, sliceLength)
	for i := range out {
		temp := binary.LittleEndian.Uint32(buf.Next(4))
		out[i] = math.Float32frombits(temp)
	}
	return out
}

//noinspection GoUnusedFunction
func interleaveFloat32(f [][]float32) []float32 {
	out := make([]float32, len(f[0])+len(f[1]))
	outIndex := 0
	for i := range f[0] {
		out[outIndex] = f[0][i]
		outIndex++
		out[outIndex] = f[1][i]
		outIndex++
	}
	return out
}

type RingBufferF struct {
	buffer    [][]float32
	readIdx   int
	writeIdx  int
	size      int
	readReset bool
}

func newRingBuffer(size int, channels int) RingBufferF {
	b := make([][]float32, channels)
	for i := 0; i < channels; i++ {
		b[i] = make([]float32, size)
	}

	readIdx := 0
	writeIdx := 0
	return RingBufferF{
		buffer:    b,
		readIdx:   readIdx,
		writeIdx:  writeIdx,
		size:      size,
		readReset: true,
	}
}

func (r *RingBufferF) writeAudio(samples [][]float32) {
	arrRemainder := r.size - r.writeIdx
	if arrRemainder < len(samples[0]) {
		for i := 0; i < len(samples); i++ {
			copy(r.buffer[i][:r.writeIdx], samples[i][:arrRemainder])
			copy(r.buffer[i], samples[i][arrRemainder:])
		}
		r.writeIdx = (len(samples[0]) + r.writeIdx) % r.size

	} else {
		for i := 0; i < len(samples); i++ {
			copy(r.buffer[i][r.writeIdx:(r.writeIdx+len(samples[0]))], samples[i])
		}
		r.writeIdx = r.writeIdx + len(samples[0])
	}
}

func (r *RingBufferF) readAudio(samples [][]float32) {
	arrRemainder := r.size - r.readIdx
	if arrRemainder < len(samples[0]) {
		for i := 0; i < len(samples); i++ {
			copy(samples[i][:arrRemainder], r.buffer[i][r.readIdx:])
			copy(samples[i][arrRemainder:], r.buffer[i])
		}
		r.readIdx = (len(samples[0]) + r.readIdx) % r.size
	} else {
		for i := 0; i < len(samples); i++ {
			copy(samples[i], r.buffer[i][r.readIdx:(r.readIdx+len(samples[0]))])
		}
		r.readIdx = r.readIdx + len(samples[0])
	}
}

func (r *RingBufferF) setLatency(bufferSamples int) {
	if r.writeIdx-bufferSamples < 0 {
		r.readIdx = (r.writeIdx - bufferSamples) + r.size
	} else {
		r.readIdx -= bufferSamples
	}
}

func (r *RingBufferF) resetReader() {
	r.readIdx = r.writeIdx
}
