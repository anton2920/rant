package main

import (
	"unsafe"
)

/* TODO(anton2920): I need to write tests for this... */
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

/* Here socket is producer and application is consumer. */
func (cb *CircularBuffer) ReadFrom(fd int32) int {
	n := int(Read(fd, cb.RemainingSlice()))
	if n > 0 {
		cb.Produce(n)
	}
	return n
}

/* TODO(anton2920): rename 'Reset.+'. */
/* ResetReader increases RemainingSpace, allowing older data to be overwritten. */
func (cb *CircularBuffer) ResetReader() {
	cb.Start = cb.Pos
	if cb.Start > len(cb.Buf)/2 {
		cb.Start -= len(cb.Buf) / 2
		cb.Pos -= len(cb.Buf) / 2
		cb.End -= len(cb.Buf) / 2
	}
}

func (cb *CircularBuffer) Consume(n int) {
	cb.Pos += n
}

func (cb *CircularBuffer) UnconsumedLen() int {
	return max(cb.End-cb.Pos, 0)
}

func (cb *CircularBuffer) UnconsumedSlice() []byte {
	return unsafe.Slice(&cb.Buf[cb.Pos], cb.UnconsumedLen())
}

func (cb *CircularBuffer) UnconsumedString() string {
	return unsafe.String(&cb.Buf[cb.Pos], cb.UnconsumedLen())
}

/* Here application is producer and socket is consumer. */
func (cb *CircularBuffer) WriteTo(fd int32) int {
	n := int(WriteFull(fd, cb.UnconsumedSlice()))
	if n > 0 {
		cb.Consume(n)
	}
	return n
}

func (cb *CircularBuffer) FullWriteTo(fd int32) int {
	ret := int(WriteFull(fd, cb.UnconsumedSlice()))
	if ret > 0 {
		cb.Start = 0
		cb.Pos = 0
		cb.End = 0
	}
	return ret
}

/* ResetWriter consumes all remaining data. */
func (cb *CircularBuffer) ResetWriter() {
	cb.Pos = cb.End
	if cb.Pos > len(cb.Buf)/2 {
		cb.Pos -= len(cb.Buf) / 2
		cb.End -= len(cb.Buf) / 2
	}
}

func (cb *CircularBuffer) Produce(n int) {
	cb.End += n
}

func (cb *CircularBuffer) RemainingSpace() int {
	return (len(cb.Buf) / 2) - (cb.End - cb.Start)
}

/* RemainingSlice returns slice of remaining free space in buffer. */
func (cb *CircularBuffer) RemainingSlice() []byte {
	return cb.Buf[cb.End : cb.Start+len(cb.Buf)/2]
}
