package main

import "unsafe"

type Buffer struct {
	Buf [2048]byte
	Pos int
	Len int
}

func (b *Buffer) FillFrom(fd int32) int {
	start := min(b.Len, b.Pos)
	if start == len(b.Buf)-1 {
		return 0
	}

	ret := int(Read(fd, b.Buf[start:]))
	if ret > 0 {
		b.Len += ret
	}
	return ret
}

func (b *Buffer) Consume(n int) {
	b.Pos += n
}

func (b *Buffer) Left() int {
	return max(b.Len-b.Pos, 0)
}

func (b *Buffer) RemainingString() string {
	return unsafe.String(&b.Buf[b.Pos], b.Left())
}

func (b *Buffer) Reset() {
	b.Pos = 0
	b.Len = 0
}
