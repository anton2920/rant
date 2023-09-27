package main

import "unsafe"

//go:noescape
//go:nosplit
func Accept(s int32, addr *SockAddr, addrlen *uint32) int32

//go:noescape
//go:nosplit
func Bind(s int32, addr *SockAddr, addrlen uint32) int32

//go:nosplit
func Close(fd int32) int32

//go:nosplit
func Exit(status int32)

//go:noescape
//go:nosplit
func Fstat(fd int32, sb *Stat) int32

//go:noescape
//go:nosplit
func Getdirentries(fd int32, buf []byte) int

//go:noescape
//go:nosplit
func Kevent(kq int32, changelist []Kevent_t, eventlist []Kevent_t, timeout *Timespec) int32

//go:nosplit
func Kqueue() int32

//go:nosplit
func Listen(s int32, backlog int32) int32

//go:nosplit
func Lseek(fd int32, offset int, whence int32) int

//go:noescape
//go:nosplit
func Nanosleep(rqtp, rmtp *Timespec) int32

//go:noescape
//go:nosplit
func Open(path string, flags int32, mode uint16) int32

//go:noescape
//go:nosplit
func Read(fd int32, buf []byte) int

//go:noescape
//go:nosplit
func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32

//go:nosplit
func Shutdown(s int32, how int32) int32

//go:nosplit
func Socket(domain, typ, protocol int32) int32

//go:noescape
//go:nosplit
func Write(fd int32, buf []byte) int
