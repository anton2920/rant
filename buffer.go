package main

import (
	"unsafe"
)

type CircularBuffer struct {
	Buf   []byte
	Start int
	Pos   int
	End   int
}

const (
	/* See <sys/mman.h>. */
	PROT_NONE  = 0x00
	PROT_READ  = 0x01
	PROT_WRITE = 0x02

	MAP_SHARED  = 0x0001
	MAP_PRIVATE = 0x0002

	MAP_FIXED = 0x0010

	MAP_ANON      = 0x1000
	MAP_ANONYMOUS = MAP_ANON
)

func (cb *CircularBuffer) FillFrom(fd int32) int {
	start := min(cb.Pos, cb.End)

	ret := int(Read(fd, cb.Buf[start:]))
	if ret > 0 {
		cb.End += ret
	}
	return ret
}

func (cb *CircularBuffer) Reset() {
	cb.Start = cb.Pos
	if cb.Start > len(cb.Buf)/2 {
		cb.Start -= len(cb.Buf) / 2
		cb.Pos -= len(cb.Buf) / 2
		cb.End -= len(cb.Buf) / 2
	}
}

func (cb *CircularBuffer) SpaceLeft() int {
	return len(cb.Buf) - (cb.End - cb.Start)
}

func (cb *CircularBuffer) Consume(n int) {
	cb.Pos += n
}

func (cb *CircularBuffer) UnconsumedString() string {
	return unsafe.String(&cb.Buf[cb.Pos], max(cb.End-cb.Pos, 0))
}

func NewCircularBuffer(size int) (*CircularBuffer, error) {
	var buffer, rb unsafe.Pointer
	var fd, ret int32

	cb := new(CircularBuffer)

	if size%int(PageSize) != 0 {
		return nil, NewError("size must be divisible by 4096", -size)
	}

	/* NOTE(anton2920): this is just (*byte)(1). */
	var SHM_ANON = unsafe.String((*byte)(unsafe.Pointer(uintptr(1))), 8)

	if fd = ShmOpen(SHM_ANON, O_RDWR, 0); fd < 0 {
		return nil, NewError("Failed to open shared memory region: ", int(fd))
	}

	if ret = Ftruncate(fd, int64(size)); ret < 0 {
		return nil, NewError("Failed to adjust size of shared memory region: ", int(ret))
	}

	if buffer = Mmap(nil, 2*uint64(size), PROT_NONE, MAP_PRIVATE|MAP_ANONYMOUS, -1, 0); buffer == nil {
		return nil, NewError("Failed to query address for future mappings: ", int(uintptr(buffer)))
	}

	if rb = Mmap(buffer, uint64(size), PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, 0); rb == nil {
		return nil, NewError("Failed to map first view of buffer: ", int(uintptr(rb)))
	}
	if Mmap(unsafe.Add(buffer, size), uint64(size), PROT_READ|PROT_WRITE, MAP_SHARED|MAP_FIXED, fd, 0) == nil {
		return nil, NewError("Failed to map second view of buffer: ", int(uintptr(rb)))
	}

	cb.Buf = unsafe.Slice((*byte)(buffer), 2*size)

	/* NOTE(anton2920): sanity checks. */
	cb.Buf[0] = '\x00'
	cb.Buf[size] = '\x00'

	return cb, nil
}
