/* TODO(anton2920):
 *
 */

package main

import "unsafe"

type SockAddrIn struct {
	Len    uint8
	Family uint8
	Port   uint16
	Addr   uint32
	_      [8]byte
}
type SockAddr = SockAddrIn

/*type SockAddr struct {
	Len    uint8
	Family uint8
	Data   [14]byte
}*/

const (
	/* NOTE(anton2920): see <sys/socket.h>. */
	AF_INET = 2
	PF_INET = AF_INET

	SOCK_STREAM = 1

	/* NOTE(anton2920): see <netinet/in.h>. */
	INADDR_ANY = 0

	SO_REUSEPORT = 0x00000200
)

func Fatal(msg string, code int32) {
	println(msg, code)
	Exit(1)
}

func Htons(x uint16) uint16 {
	/* 0xXX XX */
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func HandleConn(c int32) {
	Write(c, []byte("Hello, from the simplest Go server!\n"))
	Close(c)
}

func main() {
	const port = 7070

	var ret, l int32

	if l = Socket(PF_INET, SOCK_STREAM, 0); l < 0 {
		Fatal("Failed to create socket: ", l)
	}

	addr := SockAddrIn{Family: AF_INET, Addr: INADDR_ANY, Port: Htons(port)}
	if ret = Bind(l, &addr, uint32(unsafe.Sizeof(addr))); ret != 0 {
		Fatal("Failed to bind socket to address: ", ret)
	}

	if ret = Listen(l, 1); ret != 0 {
		Fatal("Failed to listen on the socket: ", ret)
	}
	println("Listening on 0.0.0.0:7070...")

	for {
		var c int32
		if c = Accept(l, nil, nil); c < 0 {
			println("ERROR: failed to accept connection: ", c)
			continue
		}

		go HandleConn(c)
	}
}
