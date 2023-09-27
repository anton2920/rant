package main

import "unsafe"

type SockAddrIn struct {
	Len    uint8
	Family uint8
	Port   uint16
	Addr   uint32
	_      [8]byte
}

/* NOTE(anton2920): actually SockAddr is the following structure:
 * struct sockaddr {
 *	unsigned char	sa_len;		// total length
 *	sa_family_t	sa_family;	// address family
 *	char		sa_data[14];	// actually longer; address value
 * };
 * But because I don't really care, and sizes are the same, I made them synonyms.
 */
type SockAddr = SockAddrIn

type Kevent_t struct {
	Ident  uintptr
	Filter int16
	Flags  uint16
	Fflags uint32
	Data   int
	Udata  unsafe.Pointer
	Ext    [4]uint
}

type Timespec struct {
	Sec, Nsec int
}

const (
	/* From <sys/socket.h>. */
	AF_INET = 2
	PF_INET = AF_INET

	SOCK_STREAM = 1

	SOL_SOCKET   = 0xFFFF
	SO_REUSEPORT = 0x00000200

	SHUT_WR = 1

	/* From <netinet/in.h>. */
	INADDR_ANY = 0

	/* From <fcntl.h>. */
	O_RDONLY = 0

	SEEK_SET = 0
	SEEK_END = 2

	PATH_MAX = 1024

	/* From <sys/event.h>. */
	EVFILT_VNODE = -4
	EV_ADD       = 0x0001
	EV_CLEAR     = 0x0020
	NOTE_WRITE   = 0x0002

	/* From <errno.h>. */
	EINTR = 4
)

var (
	Pages       [10][]byte
	PageKevents []Kevent_t

	IndexPage *[]byte
)

func Fatal(msg string, code int32) {
	println(msg, code)
	Exit(1)
}

func ReadFull(fd int32, buf []byte) int {
	var read, n int
	for read < len(buf) {
		if n = Read(fd, buf[read:]); n < 0 {
			if -n != EINTR {
				return n
			}
			continue
		}
		read += n
	}

	return len(buf)
}

func WriteFull(fd int32, buf []byte) int {
	var written, n int
	for written < len(buf) {
		if n = Write(fd, buf[written:]); n < 0 {
			if -n != EINTR {
				return n
			}
			continue
		}
		written += n
	}

	return len(buf)
}

func HandleConn(c int32) {
	var buffer [1024]byte

	/* NOTE(anton2920): browser must send its request first, but I don't really care about it at this point, so block until it's received. */
	Read(c, buffer[:])

	const headers = "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n"
	WriteFull(c, []byte(headers))
	WriteFull(c, *IndexPage)
	Shutdown(c, SHUT_WR)
	Close(c)
}

func ReadEntireFile(fd int32) []byte {
	var flen int
	if flen = Lseek(fd, 0, SEEK_END); flen < 0 {
		Fatal("Failed to get file length: ", int32(flen))
	}
	data := make([]byte, flen)
	if ret := Lseek(fd, 0, SEEK_SET); ret < 0 {
		Fatal("Failed to seek to the beginning of the file: ", int32(flen))
	}
	if n := ReadFull(fd, data); n < 0 {
		Fatal("Failed to read entire file: ", int32(n))
	}

	return data
}

func ReadPage(name string) *[]byte {
	var nameBuf [2 * PATH_MAX]byte
	var fd int32

	/* NOTE(anton2920): this sh**t is needed, because open(2) requires '\0'-terminated string. */
	for i := 0; i < len(name); i++ {
		nameBuf[i] = name[i]
	}
	if fd = Open(unsafe.String(&nameBuf[0], len(name)+1), O_RDONLY, 0); fd < 0 {
		Fatal("Failed to open '"+name+"': ", fd)
	}
	PageKevents = append(PageKevents, Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE})

	Pages[fd] = ReadEntireFile(fd)
	return &Pages[fd]
}

func SleepFull(time Timespec) {
	/* NOTE(anton2920): doing it in loop to fight EINTR. */
	for Nanosleep(&time, &time) < 0 {
	}
}

func MonitorPages() {
	var kq, nevents int32

	if kq = Kqueue(); kq < 0 {
		Fatal("Failed to open a kernel queue: ", kq)
	}

	if nevents = Kevent(kq, PageKevents, nil, nil); nevents < 0 {
		Fatal("Failed to register kernel events: ", nevents)
	}

	var event Kevent_t
	for {
		if nevents = Kevent(kq, nil, unsafe.Slice(&event, 1), nil); nevents < 0 {
			if -nevents != EINTR {
				println("ERROR: failed to get kernel events: ", nevents)
			}
			continue
		} else if nevents > 0 {
			println("INFO: page has been changed. Reloading...")
			Pages[event.Ident] = ReadEntireFile(int32(event.Ident))
		}

		/* NOTE(anton2920): sleep to prevent runaway events. */
		SleepFull(Timespec{Nsec: 200000000})
	}
}

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func main() {
	const port = 7070

	var ret, l int32

	IndexPage = ReadPage("pages/index.html")
	go MonitorPages()

	if l = Socket(PF_INET, SOCK_STREAM, 0); l < 0 {
		Fatal("Failed to create socket: ", l)
	}

	var enable int32 = 1
	if ret = Setsockopt(l, SOL_SOCKET, SO_REUSEPORT, unsafe.Pointer(&enable), uint32(unsafe.Sizeof(enable))); ret != 0 {
		Fatal("Failed to set socket option to reuse port: ", ret)
	}

	addr := SockAddrIn{Family: AF_INET, Addr: INADDR_ANY, Port: SwapBytesInWord(port)}
	if ret = Bind(l, &addr, uint32(unsafe.Sizeof(addr))); ret < 0 {
		Fatal("Failed to bind socket to address: ", ret)
	}

	if ret = Listen(l, 1); ret != 0 {
		Fatal("Failed to listen on the socket: ", ret)
	}
	println("Listening on 0.0.0.0:7070...")

	for {
		var c int32
		var addr SockAddrIn
		var addrLen uint32 = uint32(unsafe.Sizeof(addr))
		if c = Accept(l, &addr, &addrLen); c < 0 {
			println("ERROR: failed to accept connection: ", c)
			continue
		}
		/* println("Accepted from", addr.Addr&0xFF, (addr.Addr>>8)&0xFF, (addr.Addr>>16)&0xFF, (addr.Addr>>24)&0xFF, SwapBytesInWord(addr.Port)) */

		go HandleConn(c)
	}
}
