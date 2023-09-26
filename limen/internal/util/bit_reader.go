package util

import (
	"encoding/binary"
)

type BitReader struct {
	Data          []byte
	reg           uint64
	bytesRead     int
	totalBitsRead int
	regBits       int
	nextReg       uint64
	nextRegBits   int
}

func (r *BitReader) ReadBits(bits int, payload *uint64) bool {
	if bits == 0 {
		*payload = 0
		return true
	}

	if bits > r.regBits && !r.refillReg(bits) {
		r.regBits = 0
		r.reg = 0
		*payload = 0
		return false
	}

	if bits == 64 {
		*payload = r.reg
		r.reg = 0
		r.regBits = 0
		return true
	}

	r.totalBitsRead += bits

	*payload = r.reg >> (64 - bits)
	r.reg <<= bits
	r.regBits -= bits
	return true
}

func (r *BitReader) SkipBits(bits int) bool {
	totalRegBits := r.regBits + r.nextRegBits
	if totalRegBits >= bits {
		return r.smallSkipBits(bits)
	}

	// skip bits stored in registers
	bits -= totalRegBits
	r.totalBitsRead += totalRegBits
	r.regBits = 0
	r.reg = 0
	r.nextRegBits = 0
	r.nextReg = 0

	// skip full bytes if any
	bytes := bits / 8

	if bytes > 0 {
		if r.bytesRead+bytes > len(r.Data) {
			return false
		}

		r.bytesRead += bytes
		bits -= 8 * bytes
		r.totalBitsRead += 8 * bytes
	}

	return r.smallSkipBits(bits)
}

func (r *BitReader) BitsAvailable() int {
	return 8*len(r.Data) - r.totalBitsRead
}

func (r *BitReader) ReadSlice(payload []byte) bool {
	available := r.BitsAvailable()
	bytesToRead := len(payload)
	if 8*bytesToRead > available || available%8 != 0 {
		return false
	}

	var readPayload uint64

	idx := 0
	shift := 56
	for bytesToRead > 0 && r.regBits > 0 {
		payload[idx] = byte(r.reg >> (64 - shift))
		idx += 1
		shift -= 8
		r.regBits -= 8
		r.totalBitsRead += 8
		r.bytesRead += 1
		bytesToRead -= 1

	}

	shift = 56
	for bytesToRead > 0 && r.nextRegBits > 0 {
		payload[idx] = byte(r.nextReg >> (64 - shift))
		idx += 1
		shift -= 8
		r.nextRegBits -= 8
		r.totalBitsRead += 8
		r.bytesRead += 1
		bytesToRead -= 1
	}

	for bytesToRead > 0 {
		payload[idx] = r.Data[r.bytesRead]
		r.bytesRead += 1
		r.totalBitsRead += 8
		bytesToRead -= 1
	}

	for bytesToRead > 8 {
		r.ReadBits(64, &readPayload)
	}

	return true
}

func (r *BitReader) smallSkipBits(bits int) bool {
	var dummy uint64
	for bits >= 64 {
		if !r.ReadBits(64, &dummy) {
			return false
		}
		bits -= 64
		r.totalBitsRead += 64
	}
	return r.ReadBits(bits, &dummy)
}

func (r *BitReader) refillReg(minBits int) bool {
	r.refillCurrentReg()
	if minBits <= r.regBits {
		return true
	}

	const RegByteSize = 8

	if r.bytesRead >= len(r.Data) {
		return false
	}

	readTo := r.bytesRead + RegByteSize
	if readTo > len(r.Data) {
		readTo = len(r.Data)
	}

	toRead := readTo - r.bytesRead

	tmpReg := [8]byte{}
	copy(tmpReg[:], r.Data[r.bytesRead:readTo])

	r.nextReg = binary.BigEndian.Uint64(tmpReg[:])
	r.nextRegBits = toRead * 8
	r.bytesRead = readTo

	r.refillCurrentReg()

	return r.regBits >= minBits
}

func (r *BitReader) refillCurrentReg() {
	// if either current register is full or next is empty we can refill anything
	if r.regBits == 64 || r.nextRegBits == 0 {
		return
	}

	r.reg |= (r.nextReg >> r.regBits)
	freeRegBits := 64 - r.regBits
	if freeRegBits >= r.nextRegBits {
		r.regBits += r.nextRegBits
		r.nextReg = 0
		r.nextRegBits = 0
		return
	}

	r.regBits += freeRegBits
	r.nextReg <<= uint64(freeRegBits)
	r.nextRegBits -= freeRegBits
}
