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

const (
	/* NOTE(anton2920): see <sys/socket.h>. */
	AF_INET = 2
	PF_INET = AF_INET

	SOCK_STREAM = 1

	SOL_SOCKET   = 0xFFFF
	SO_REUSEPORT = 0x00000200

	SHUT_WR = 1

	/* NOTE(anton2920): see <netinet/in.h>. */
	INADDR_ANY = 0

	/* NOTE(anton2920): see <fcntl.h>. */
	O_RDONLY = 0

	SEEK_SET = 0
	SEEK_END = 2
)

var (
	IndexPage []byte
)

func Fatal(msg string, code int32) {
	println(msg, code)
	Exit(1)
}

func WriteAll(c int32, buf []byte) {
	var sent int
	for sent < len(buf) {
		n := Write(c, buf[sent:])
		sent += n
	}
}

func HandleConn(c int32) {
	var buffer [1024]byte

	/* NOTE(anton2920): browser must send its request first, but I don't really care about it at this point, so block until it's received. */
	Read(c, buffer[:])

	const headers = "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n"
	WriteAll(c, []byte(headers))
	WriteAll(c, IndexPage)
	Shutdown(c, SHUT_WR)
	Close(c)
}

/* NOTE(anton2920): string must be '\0'-terminated. */
func ReadPage(name string, data *[]byte) {
	var fd int32
	var flen int

	if fd = Open(name, O_RDONLY, 0); fd < 0 {
		Fatal("Failed to open '"+name+"': ", fd)
	}
	if flen = Lseek(fd, 0, SEEK_END); flen < 0 {
		Fatal("Failed to get file length: ", int32(flen))
	}
	*data = make([]byte, flen)
	if ret := Lseek(fd, 0, SEEK_SET); ret < 0 {
		Fatal("Failed to seek to the beginning of the file: ", int32(flen))
	}
	if n := Read(fd, *data); n != flen {
		Fatal("Failed to read page contents: ", int32(n))
	}
	Close(fd)
}

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func main() {
	const port = 7070

	var ret, l int32

	ReadPage("pages/index.html\x00", &IndexPage)

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
