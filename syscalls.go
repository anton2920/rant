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
func Bind(s int32, addr *SockAddr, addrlen uint32) int32

//go:noescape
//go:nosplit
func ClockGettime(clockID int32, tp *Timespec) int32

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
func Fstat(fd int32, sb *Stat) int32

//go:noescape
func Ftruncate(fd int32, length int64) int32

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
func Kqueue() int32

//go:nosplit
func Listen(s int32, backlog int32) int32

//go:nosplit
func Lseek(fd int32, offset int64, whence int32) int64

//go:noescape
//go:nosplit
func Mmap(addr unsafe.Pointer, len uint64, prot, flags, fd int32, offset int64) unsafe.Pointer

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
func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32

//go:noescape
//go:nosplit
func ShmOpen2(path string, flags int32, mode uint16, shmflags int32, name unsafe.Pointer) int32

//go:nosplit
func Shutdown(s int32, how int32) int32

//go:nosplit
func Socket(domain, typ, protocol int32) int32

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
