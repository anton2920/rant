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

/* See <sys/stat.h>. */
type Stat struct {
	Dev       uint   /* inode's device */
	Ino       uint   /* inode's number */
	Nlink     uint64 /* number of hard links */
	Mode      uint16 /* inode protection mode */
	_         int16
	Uid       uint32 /* user ID of the file's owner */
	Gid       uint32 /* group ID of the file's group */
	_         int32
	Rdev      uint64   /* device type */
	Atime     Timespec /* time of last access */
	Mtime     Timespec /* time of last data modification */
	Ctime     Timespec /* time of last file status change */
	Birthtime Timespec /* time of file creation */
	Size      int      /* file size, in bytes */
	Blocks    int      /* blocks allocated for file */
	Blksize   int32    /* optimal blocksize for I/O */
	Flags     uint32   /* user defined flags for file */
	Gen       uint64   /* file generation number */
	_         [10]int
}

type Request struct {
	Method string
	Path   string
	Query  string
}

type Tweet struct {
	Ctime int
	Text  []byte
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
	ENOENT = 2
	EINTR  = 4
)

const (
	ResponseOK         = "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n"
	ResponseBadRequest = "HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>400 Bad Request</title></head><body><h1>400 Bad Request</h1><p>Your browser sent a request that this server could not understand.</p></body></html>"
	ResponseNotFound   = "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p>The requested URL was not found on this server.</p></body></html>"
)

var (
	Pages       [10][]byte
	PageKevents []Kevent_t

	IndexPage *[]byte
	TweetPage *[]byte

	Tweets []Tweet
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

func IndexPageHandler(c int32, r *Request) {
	var buffer [100]byte

	if r.Query != "" {
		WriteFull(c, []byte(r.Query))
	}

	const tweetBeforeID = `<div class="tweet"><a href="/tweet/`
	const tweetBeforeDate = `"><div class="tweet-insides"><img class="tweet-avatar" src="https://media.licdn.com/dms/image/C4E03AQGi1v1OmgpUTQ/profile-displayphoto-shrink_800_800/0/1600259320098?e=1701302400&v=beta&t=SohoOoRvVqYuyUE7QnPQWYb-8Tm-Yc6ZUA75Wd_s2-4" alt="Profile picture"><div><div class="tweet-header"><b>Anton Pavlovskii</b><span>@anton2920 `
	const tweetBeforeText = `</span></div><p>`
	const tweetAfterText = `</p></div></div></a></div>`
	const finisher = `</body></html>`

	WriteFull(c, []byte(ResponseOK))
	WriteFull(c, *IndexPage)
	for i := len(Tweets) - 1; i >= 0; i-- {
		tweet := Tweets[i]

		ndigits := SlicePutInt(unsafe.Slice(&buffer[0], len(buffer)), i)

		WriteFull(c, []byte(tweetBeforeID))
		WriteFull(c, unsafe.Slice(&buffer[0], ndigits))
		WriteFull(c, []byte(tweetBeforeDate))
		/* TODO(anton2920): insert date. */
		WriteFull(c, []byte(tweetBeforeText))
		WriteFull(c, tweet.Text)
		WriteFull(c, []byte(tweetAfterText))
	}
	WriteFull(c, []byte(finisher))
}

func StrToInt(xs string) (int, bool) {
	var sign bool = true
	var ret int

	for _, x := range xs {
		if x == '-' {
			sign = false
		} else if (x < '0') || (x > '9') {
			return 0, false
		}
		ret = (ret * 10) + int(x-'0')
	}

	if !sign {
		ret = -ret
	}

	return ret, true
}

func TweetPageHandler(c int32, r *Request) {
	const tweetBeforeDate = `<div class="tweet"><div class="tweet-insides"><img class="tweet-avatar" src="https://media.licdn.com/dms/image/C4E03AQGi1v1OmgpUTQ/profile-displayphoto-shrink_800_800/0/1600259320098?e=1701302400&v=beta&t=SohoOoRvVqYuyUE7QnPQWYb-8Tm-Yc6ZUA75Wd_s2-4" alt="Profile picture"><div><div class="tweet-header"><b>Anton Pavlovskii</b><span>@anton2920 `
	const tweetBeforeText = `</span></div><p>`
	const finisher = `</p></div></div></div></body></html>`

	id, ok := StrToInt(r.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(Tweets)-1) {
		WriteFull(c, []byte(ResponseNotFound))
		return
	}
	tweet := Tweets[id]

	WriteFull(c, []byte(ResponseOK))
	WriteFull(c, *TweetPage)
	WriteFull(c, []byte(tweetBeforeDate))
	/* TODO(anton2920): insert date. */
	WriteFull(c, []byte(tweetBeforeText))
	WriteFull(c, tweet.Text)
	WriteFull(c, []byte(finisher))
}

func HandleConn(c int32) {
	var buffer [512]byte
	Read(c, unsafe.Slice(&buffer[0], len(buffer)))

	var r Request
	if unsafe.String(&buffer[0], 3) == "GET" {
		r.Method = "GET"

		lineEnd := FindRune(unsafe.String(&buffer[len(r.Method)+1], len(buffer)-len(r.Method)+1), '\r')
		requestLine := unsafe.Slice(&buffer[len(r.Method)+1], lineEnd-1) /* without method. */

		pathEnd := FindRune(unsafe.String(&requestLine[0], len(requestLine)), '?')
		if pathEnd != -1 {
			/* With query. */
			r.Path = unsafe.String(&requestLine[0], pathEnd)

			queryStart := pathEnd + 1
			queryEnd := FindRune(unsafe.String(&requestLine[queryStart], len(requestLine)-queryStart), ' ')
			r.Query = unsafe.String(&requestLine[queryStart], queryEnd)
		} else {
			/* No query. */
			pathEnd = FindRune(unsafe.String(&requestLine[0], len(requestLine)), ' ')
			r.Path = unsafe.String(&requestLine[0], pathEnd)
		}
	} else {
		WriteFull(c, []byte(ResponseBadRequest))
		Close(c)
		return
	}
	/* println(r.Method, len(r.Method), r.Path, len(r.Path), r.Query, len(r.Query)) */

	if r.Path == "/" {
		IndexPageHandler(c, &r)
	} else if (len(r.Path) == len("/favicon.ico")) && (r.Path == "/favicon.ico") {
		/* Do nothing :) */
	} else if (len(r.Path) > len("/tweet/")) && (r.Path[:len("/tweet/")] == "/tweet/") {
		TweetPageHandler(c, &r)
	} else {
		WriteFull(c, []byte(ResponseNotFound))
	}
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

func SlicePutInt(buf []byte, x int) int {
	var sign bool
	var rx int

	if x == 0 {
		buf[0] = '0'
		return 1
	}

	sign = (x > 0)

	for x > 0 {
		rx = (10 * rx) + (x % 10)
		x /= 10
	}

	var i int
	if !sign {
		buf[i] = '-'
		i++
	}
	for ; rx > 0; i++ {
		buf[i] = byte((rx % 10) + '0')
		rx /= 10
	}
	return i
}

func ReadTweets() {
	const tweetsPath = "tweets/"

	var buffer [PATH_MAX]byte
	var fd int32
	var st Stat

	copy(buffer[:], []byte(tweetsPath))

	var tweet Tweet
	for i := 0; ; i++ {
		SlicePutInt(buffer[len(tweetsPath):], i)

		if fd = Open(unsafe.String(&buffer[0], len(buffer)), O_RDONLY, 0); fd < 0 {
			if -fd != ENOENT {
				Fatal("Failed to open '"+string(buffer[:])+"': ", fd)
			}
			return
		}
		if ret := Fstat(fd, &st); ret < 0 {
			Fatal("Failed to get stat of '"+string(buffer[:])+"': ", ret)
		}

		tweet.Ctime = st.Ctime.Sec
		tweet.Text = ReadEntireFile(fd)
		Close(fd)

		Tweets = append(Tweets, tweet)
	}
}

func SwapBytesInWord(x uint16) uint16 {
	return ((x << 8) & 0xFF00) | (x >> 8)
}

func main() {
	const port = 7070

	var ret, l int32

	IndexPage = ReadPage("pages/index.html")
	TweetPage = ReadPage("pages/tweet.html")
	go MonitorPages()

	ReadTweets()

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
