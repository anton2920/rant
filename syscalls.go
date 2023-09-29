package main

import "unsafe"

//go:noescape
//go:nosplit
func RawAccept(s int32, addr *SockAddr, addrlen *uint32) int32

func Accept(s int32, addr *SockAddr, addrlen *uint32) (ret int32) {
	SyscallEnter()
	ret = RawAccept(s, addr, addrlen)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawBind(s int32, addr *SockAddr, addrlen uint32) int32

func Bind(s int32, addr *SockAddr, addrlen uint32) (ret int32) {
	SyscallEnter()
	ret = RawBind(s, addr, addrlen)
	SyscallExit()
	return
}

//go:nosplit
func RawClose(fd int32) int32

func Close(fd int32) (ret int32) {
	SyscallEnter()
	ret = RawClose(fd)
	SyscallExit()
	return
}

//go:nosplit
func Exit(status int32)

//go:noescape
//go:nosplit
func RawFstat(fd int32, sb *Stat) int32

func Fstat(fd int32, sb *Stat) (ret int32) {
	SyscallEnter()
	ret = RawFstat(fd, sb)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawKevent(kq int32, changelist []Kevent_t, eventlist []Kevent_t, timeout *Timespec) int32

func Kevent(kq int32, changelist []Kevent_t, eventlist []Kevent_t, timeout *Timespec) (ret int32) {
	SyscallEnter()
	ret = RawKevent(kq, changelist, eventlist, timeout)
	SyscallExit()
	return
}

//go:nosplit
func RawKqueue() int32

func Kqueue() (ret int32) {
	SyscallEnter()
	ret = RawKqueue()
	SyscallExit()
	return
}

//go:nosplit
func RawListen(s int32, backlog int32) int32

func Listen(s int32, backlog int32) (ret int32) {
	SyscallEnter()
	ret = RawListen(s, backlog)
	SyscallExit()
	return
}

//go:nosplit
func RawLseek(fd int32, offset int64, whence int32) int64

func Lseek(fd int32, offset int64, whence int32) (ret int64) {
	SyscallEnter()
	ret = RawLseek(fd, offset, whence)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawNanosleep(rqtp, rmtp *Timespec) int32

func Nanosleep(rqtp, rmtp *Timespec) (ret int32) {
	SyscallEnter()
	ret = RawNanosleep(rqtp, rmtp)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawOpen(path string, flags int32, mode uint16) int32

func Open(path string, flags int32, mode uint16) (ret int32) {
	SyscallEnter()
	ret = RawOpen(path, flags, mode)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawRead(fd int32, buf []byte) int64

func Read(fd int32, buf []byte) (ret int64) {
	SyscallEnter()
	ret = RawRead(fd, buf)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawSetsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32

func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) (ret int32) {
	SyscallEnter()
	ret = RawSetsockopt(s, level, optname, optval, optlen)
	SyscallExit()
	return
}

//go:nosplit
func RawShutdown(s int32, how int32) int32

func Shutdown(s int32, how int32) (ret int32) {
	SyscallEnter()
	ret = RawShutdown(s, how)
	SyscallExit()
	return
}

//go:nosplit
func RawSocket(domain, typ, protocol int32) int32

func Socket(domain, typ, protocol int32) (ret int32) {
	SyscallEnter()
	ret = RawSocket(domain, typ, protocol)
	SyscallExit()
	return
}

//go:noescape
//go:nosplit
func RawWrite(fd int32, buf []byte) int64

func Write(fd int32, buf []byte) (ret int64) {
	SyscallEnter()
	ret = RawWrite(fd, buf)
	SyscallExit()
	return
}

//go:linkname SyscallEnter runtime.entersyscall
func SyscallEnter()

//go:linkname SyscallExit runtime.exitsyscall
func SyscallExit()
