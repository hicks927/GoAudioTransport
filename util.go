package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
		out[i] = float32(binary.LittleEndian.Uint32(buf.Next(4)))
	}
	return out
}

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
