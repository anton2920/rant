package main

import (
	"unsafe"
)

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

/* See <time.h>. */
type Tm struct {
	Sec   int /* seconds after the minute [0-60] */
	Min   int /* minutes after the hour [0-59] */
	Hour  int /* hours since midnight [0-23] */
	Mday  int /* day of the month [1-31] */
	Mon   int /* months since January [0-11] */
	Year  int /* years since 1900 */
	Wday  int /* days since Sunday [0-6] */
	Yday  int /* days since January 1 [0-365] */
	Isdst int /* Daylight Savings Time flag */
}

type Request struct {
	Method string
	Path   string
	Query  string
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
	ResponseFinisher   = "</div></div></div></body></html>"
)

var (
	Pages       [10][]byte
	PageKevents []Kevent_t

	IndexPage *[]byte
	TweetPage *[]byte

	TweetHTMLs [][]byte
	TweetTexts [][]byte

	IndexPageFull []byte
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

func TimeToTm(t int) Tm {
	var tm Tm

	daysSinceJan1st := [2][13]int{
		{0, 31, 59, 90, 120, 151, 181, 212, 243, 273, 304, 334, 365}, // 365 days, non-leap
		{0, 31, 60, 91, 121, 152, 182, 213, 244, 274, 305, 335, 366}, // 366 days, leap
	}

	t += 3 * 60 * 60 /* MSK timezone hack. */

	/* Re-bias from 1970 to 1601: 1970 - 1601 = 369 = 3*100 + 17*4 + 1 years (incl. 89 leap days) = (3*100*(365+24/100) + 17*4*(365+1/4) + 1*365)*24*3600 seconds. */
	sec := t + 11644473600

	wday := (sec/86400 + 1) % 7 /* day of week */

	/* Remove multiples of 400 years (incl. 97 leap days). */
	quadricentennials := sec / 12622780800 /* 400*365.2425*24*3600 .*/
	sec %= 12622780800

	/* Remove multiples of 100 years (incl. 24 leap days), can't be more than 3 (because multiples of 4*100=400 years (incl. leap days) have been removed). */
	centennials := sec / 3155673600 /* 100*(365+24/100)*24*3600. */
	if centennials > 3 {
		centennials = 3
	}
	sec -= centennials * 3155673600

	/* Remove multiples of 4 years (incl. 1 leap day), can't be more than 24 (because multiples of 25*4=100 years (incl. leap days) have been removed). */
	quadrennials := sec / 126230400 /*  4*(365+1/4)*24*3600. */
	if quadrennials > 24 {
		quadrennials = 24
	}
	sec -= quadrennials * 126230400

	/* Remove multiples of years (incl. 0 leap days), can't be more than 3 (because multiples of 4 years (incl. leap days) have been removed). */
	annuals := sec / 31536000 // 365*24*3600
	if annuals > 3 {
		annuals = 3
	}
	sec -= annuals * 31536000

	/* Calculate the year and find out if it's leap. */
	year := 1601 + quadricentennials*400 + centennials*100 + quadrennials*4 + annuals
	var leap int
	if (year%4 == 0) && ((year%100 != 0) || (year%400 == 0)) {
		leap = 1
	} else {
		leap = 0
	}

	/* Calculate the day of the year and the time. */
	yday := sec / 86400
	sec %= 86400
	hour := sec / 3600
	sec %= 3600
	min := sec / 60
	sec %= 60

	/* Calculate the month. */
	var month, mday int = 1, 1
	for ; month < 13; month++ {
		if yday < daysSinceJan1st[leap][month] {
			mday += yday - daysSinceJan1st[leap][month-1]
			break
		}
	}

	tm.Sec = sec          /*  [0,59]. */
	tm.Min = min          /*  [0,59]. */
	tm.Hour = hour        /*  [0,23]. */
	tm.Mday = mday        /*  [1,31]  (day of month). */
	tm.Mon = month - 1    /*  [0,11]  (month). */
	tm.Year = year - 1900 /*  70+     (year since 1900). */
	tm.Wday = wday        /*  [0,6]   (day since Sunday AKA day of week). */
	tm.Yday = yday        /*  [0,365] (day since January 1st AKA day of year). */
	tm.Isdst = -1         /*  daylight saving time flag. */

	return tm
}

func SlicePutTm(buf []byte, tm Tm) int {
	var n, ndigits int

	if tm.Mday+1 < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Mday)
	n += ndigits
	buf[n] = '.'
	n++

	if tm.Mon+1 < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Mon+1)
	n += ndigits
	buf[n] = '.'
	n++

	ndigits = SlicePutPositiveInt(buf[n:], tm.Year+1900)
	n += ndigits
	buf[n] = ' '
	n++

	if tm.Hour < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Hour)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Min < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Min)
	n += ndigits
	buf[n] = ':'
	n++

	if tm.Sec < 10 {
		buf[n] = '0'
		n++
	}
	ndigits = SlicePutPositiveInt(buf[n:], tm.Sec)
	n += ndigits
	buf[n] = ' '
	n++

	buf[n] = 'M'
	buf[n+1] = 'S'
	buf[n+2] = 'K'

	return n + 3
}

/* CharToByte returns ASCII-decoded character. For example, 'A' yields '\x0A'. */
func CharToByte(c byte) (byte, bool) {
	if c >= '0' && c <= '9' {
		return c - '0', true
	} else if c >= 'A' && c <= 'F' {
		return 10 + c - 'A', true
	} else {
		return '\x00', false
	}
}

func URLDecode(decoded []byte, encoded string) (int, bool) {
	var hi, lo byte
	var ok bool
	var n int

	for i := 0; i < len(encoded); i++ {
		if encoded[i] == '%' {
			hi = encoded[i+1]
			hi, ok = CharToByte(hi)
			if !ok {
				return 0, false
			}

			lo = encoded[i+2]
			lo, ok = CharToByte(lo)
			if !ok {
				return 0, false
			}

			decoded[n] = byte(hi<<4 | lo)
			i += 2
		} else if encoded[i] == '+' {
			decoded[n] = ' '
		} else {
			decoded[n] = encoded[i]
		}
		n++
	}
	return n, true
}

func IndexPageHandler(c int32, r *Request) {
	const maxQueryLen = 1024
	var queryString string

	if r.Query != "" {
		if r.Query[:len("Query=")] != "Query=" {
			WriteFull(c, []byte(ResponseBadRequest))
			return
		}

		queryString = r.Query[len("Query="):]
		if len(queryString) > maxQueryLen {
			WriteFull(c, []byte(ResponseBadRequest))
			return
		}
	}

	if queryString != "" {
		var decodedQuery [maxQueryLen]byte
		decodedLen, ok := URLDecode(unsafe.Slice(&decodedQuery[0], len(decodedQuery)), queryString)
		if !ok {
			WriteFull(c, []byte(ResponseBadRequest))
			return
		}

		WriteFull(c, []byte(ResponseOK))
		WriteFull(c, *IndexPage)
		/* WriteFull(c, unsafe.Slice(&decodedQuery[0], decodedLen)) */
		for i := len(TweetHTMLs) - 1; i >= 0; i-- {
			if FindSubstring(unsafe.String(unsafe.SliceData(TweetTexts[i]), len(TweetTexts[i])), unsafe.String(&decodedQuery[0], decodedLen)) != -1 {
				WriteFull(c, TweetHTMLs[i])
			}
		}
	} else {
		WriteFull(c, []byte(ResponseOK))
		WriteFull(c, IndexPageFull)
	}
	WriteFull(c, []byte(ResponseFinisher))
}

func StrToPositiveInt(xs string) (int, bool) {
	var ret int

	for _, x := range xs {
		if (x < '0') || (x > '9') {
			return 0, false
		}
		ret = (ret * 10) + int(x-'0')
	}

	return ret, true
}

func TweetPageHandler(c int32, r *Request) {
	id, ok := StrToPositiveInt(r.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(TweetHTMLs)-1) {
		WriteFull(c, []byte(ResponseNotFound))
		return
	}

	WriteFull(c, []byte(ResponseOK))
	WriteFull(c, *TweetPage)
	WriteFull(c, TweetHTMLs[id])
	WriteFull(c, []byte(ResponseFinisher))
}

func HandleConn(c int32) {
	var buffer [512]byte
	Read(c, unsafe.Slice(&buffer[0], len(buffer)))

	var r Request
	if unsafe.String(&buffer[0], 3) == "GET" {
		r.Method = "GET"

		lineEnd := FindRune(unsafe.String(&buffer[len(r.Method)+1], len(buffer)-len(r.Method)+1), '\r')
		requestLine := unsafe.String(&buffer[len(r.Method)+1], lineEnd-1) /* without method. */

		pathEnd := FindRune(requestLine, '?')
		if pathEnd != -1 {
			/* With query. */
			r.Path = unsafe.String(unsafe.StringData(requestLine), pathEnd)

			queryStart := pathEnd + 1
			queryEnd := FindRune(unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(unsafe.StringData(requestLine)), queryStart)), len(requestLine)-queryStart), ' ')
			r.Query = unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(unsafe.StringData(requestLine)), queryStart)), queryEnd)
		} else {
			/* No query. */
			pathEnd = FindRune(requestLine, ' ')
			r.Path = unsafe.String(unsafe.StringData(requestLine), pathEnd)
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

func SlicePutPositiveInt(buf []byte, x int) int {
	var ndigits int
	var rx int

	if x == 0 {
		buf[0] = '0'
		return 1
	}

	for x > 0 {
		rx = (10 * rx) + (x % 10)
		x /= 10
		ndigits++
	}

	var i int
	for i = 0; ndigits > 0; i++ {
		buf[i] = byte((rx % 10) + '0')
		rx /= 10
		ndigits--
	}
	return i
}

func ReadTweets() {
	const tweetsPath = "tweets/"

	const tweetBeforeDate = `<div class="tweet"><div class="tweet-insides"><img class="tweet-avatar" src="https://media.licdn.com/dms/image/C4E03AQGi1v1OmgpUTQ/profile-displayphoto-shrink_800_800/0/1600259320098?e=1701302400&v=beta&t=SohoOoRvVqYuyUE7QnPQWYb-8Tm-Yc6ZUA75Wd_s2-4" alt="Profile picture"><div><div class="tweet-header"><a href="/"><b>Anton Pavlovskii</b><span>@anton2920 `
	const tweetBeforeID = `</span></a></div><a href="/tweet/`
	const tweetBeforeText = `"><p>`
	const tweetAfterText = `</p></div></div></a></div>`

	var pathBuf [PATH_MAX]byte

	var idBuf [10]byte
	var idBufLen int

	var dateBuf [25]byte
	var dateBufLen int

	var fd int32
	var st Stat

	copy(unsafe.Slice(&pathBuf[0], len(pathBuf)), []byte(tweetsPath))

	TweetHTMLs = make([][]byte, 0, 128)

	for i := 0; ; i++ {
		tweet := make([]byte, 0, 2048)

		idBufLen = SlicePutPositiveInt(unsafe.Slice(&idBuf[0], len(idBuf)), i)
		copy(unsafe.Slice(&pathBuf[len(tweetsPath)], len(pathBuf)-len(tweetsPath)), unsafe.Slice(&idBuf[0], idBufLen))

		if fd = Open(unsafe.String(&pathBuf[0], len(pathBuf)), O_RDONLY, 0); fd < 0 {
			if -fd != ENOENT {
				Fatal("Failed to open '"+string(pathBuf[:])+"': ", fd)
			}
			return
		}
		if ret := Fstat(fd, &st); ret < 0 {
			Fatal("Failed to get stat of '"+string(pathBuf[:])+"': ", ret)
		}
		dateBufLen = SlicePutTm(unsafe.Slice(&dateBuf[0], len(dateBuf)), TimeToTm(st.Birthtime.Sec))

		text := ReadEntireFile(fd)
		TweetTexts = append(TweetTexts, text)
		Close(fd)

		tweet = append(tweet, tweetBeforeDate...)
		tweet = append(tweet, unsafe.Slice(&dateBuf[0], dateBufLen)...)
		tweet = append(tweet, tweetBeforeID...)
		tweet = append(tweet, unsafe.Slice(&idBuf[0], idBufLen)...)
		tweet = append(tweet, tweetBeforeText...)
		tweet = append(tweet, text...)
		tweet = append(tweet, tweetAfterText...)

		TweetHTMLs = append(TweetHTMLs, tweet)
	}
}

func ConstructIndexPage() {
	IndexPageFull = make([]byte, 0, 4<<10)

	IndexPageFull = append(IndexPageFull, *IndexPage...)
	for i := len(TweetHTMLs) - 1; i >= 0; i-- {
		IndexPageFull = append(IndexPageFull, TweetHTMLs[i]...)
	}
}

func MonitorTweets() {
	var fd, kq, nevents int32

	if kq = Kqueue(); kq < 0 {
		Fatal("Failed to open a kernel queue: ", kq)
	}

	const tweetsDir = "./tweets\x00"
	if fd = Open(tweetsDir, O_RDONLY, 0); fd < 0 {
		Fatal("Failed to open '"+tweetsDir+"': ", fd)
	}

	tweetsKevent := Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE}
	if nevents = Kevent(kq, unsafe.Slice(&tweetsKevent, 1), nil, nil); nevents < 0 {
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
			println("INFO: change in tweets directory. Reloading...")
			ReadTweets()
			ConstructIndexPage()
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
	TweetPage = ReadPage("pages/tweet.html")

	ReadTweets()
	ConstructIndexPage()

	go MonitorPages()
	go MonitorTweets()

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
