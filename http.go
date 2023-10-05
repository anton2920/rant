package main

import (
	"unsafe"
)

type Request struct {
	Method  string
	URL     URL
	Version string
}

type Response struct {
	Code        int
	ContentType string
	Body        []byte
}

type HTTPRouter func(w *Response, r *Request)

const (
	StatusOK         = 200
	StatusBadRequest = 400
	StatusNotFound   = 404
)

const (
	ResponseBadRequest            = "HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\nContent-Length: 175\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>400 Bad Request</title></head><body><h1>400 Bad Request</h1><p>Your browser sent a request that this server could not understand.</p></body></html>"
	ResponseNotFound              = "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nContent-Length: 152\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p>The requested URL was not found on this server.</p></body></html>"
	ResponseMethodNotAllowed      = "HTTP/1.1 405 Method Not Allowed\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	ResponseRequestTimeout        = "HTTP/1.1 408 Request Timeout\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	ResponseRequestEntityTooLarge = "HTTP/1.1 413 Request Entity Too Large\r\nContent-Type: text/html\r\nConent-Length: ...\r\nConnection: close\r\n\r\n"
)

const (
	HeaderContentLength = `Content-Length: `
)

const (
	STATE_UNKNOWN = iota
	STATE_NEED_MORE
	STATE_NEED_FLUSH

	STATE_METHOD
	STATE_URI
	STATE_VERSION
	STATE_HEADER
	STATE_BODY

	STATE_BAD_REQUEST
	STATE_METHOD_NOT_ALLOWED
	STATE_REQUEST_TIMEOUT
	STATE_REQUEST_ENTITY_TOO_LARGE

	STATE_DONE
)

func HTTPWorker(cc <-chan int32, router HTTPRouter) {
	const maxResponseOffset = 128
	const pipelining = true

	var err error

	var reqB *CircularBuffer
	var r Request

	var contentLengthBuf []byte
	var respB *CircularBuffer
	var w Response

	var currState, prevState int

	contentLengthBuf = make([]byte, len(HeaderContentLength)+10)
	copy(contentLengthBuf, []byte(HeaderContentLength))

	if reqB, err = NewCircularBuffer(10 * PageSize); err != nil {
		FatalError(err)
	}

	if respB, err = NewCircularBuffer(1024 * PageSize); err != nil {
		FatalError(err)
	}

	w.Body = make([]byte, 64*1024)

	for c := range cc {
		respB.ResetWriter()
		respB.ResetReader()

	connectionFor:
		for {
			reqB.ResetReader()
			currState = STATE_METHOD

			if pipelining {
				w.Body = w.Body[:0]
			} else {
				w.Body = respB.RemainingSlice()[maxResponseOffset:maxResponseOffset]
			}
			w.ContentType = ""
			w.Code = 0

			for currState != STATE_DONE {
				switch currState {
				case STATE_UNKNOWN:
					remaining := reqB.UnconsumedString()
					if len(remaining) < 2 {
						prevState = STATE_UNKNOWN
						currState = STATE_NEED_MORE
						continue
					}
					if remaining[:2] == "\r\n" {
						reqB.Consume(len("\r\n"))
						currState = STATE_DONE
					} else {
						currState = STATE_HEADER
					}

				case STATE_NEED_MORE:
					if reqB.RemainingSpace() == 0 {
						currState = STATE_REQUEST_ENTITY_TOO_LARGE
						continue
					}

					if pipelining && respB.UnconsumedLen() > 0 {
						currState = STATE_NEED_FLUSH
						continue
					}

					n := reqB.ReadFrom(c)
					if n <= 0 {
						switch -n {
						case 0:
							/* End of file. */
						case ECONNRESET:
							/* Connection reset by peer. */
						case EWOULDBLOCK:
							if prevState != STATE_METHOD {
								prevState = STATE_REQUEST_TIMEOUT
							}
							Shutdown(c, SHUT_RD)
						default:
							println("ERROR: failed to read buffer: ", n)
						}
						break connectionFor
					}
					currState = prevState
					prevState = STATE_UNKNOWN
				case STATE_NEED_FLUSH:
					n := respB.FullWriteTo(c)
					if n <= 0 {
						switch -n {
						case 0:
							/* End of file. */
						case EPIPE:
							/* Broken pipe. */
						case ECONNRESET:
							/* Connection reset by peer */
						default:
							println("ERROR: failed to write buffer: ", n)
						}
						break connectionFor
					}
					currState = prevState
					prevState = STATE_UNKNOWN
				case STATE_METHOD:
					remaining := reqB.UnconsumedString()
					if len(remaining) < 3 {
						prevState = STATE_METHOD
						currState = STATE_NEED_MORE
						continue
					}
					switch remaining[:3] {
					case "GET":
						r.Method = "GET"
					default:
						currState = STATE_METHOD_NOT_ALLOWED
						continue
					}
					reqB.Consume(len(r.Method) + 1)
					currState = STATE_URI
				case STATE_URI:
					remaining := reqB.UnconsumedString()
					lineEnd := FindChar(remaining, '\r')
					if lineEnd == -1 {
						prevState = STATE_URI
						currState = STATE_NEED_MORE
						continue
					}

					uriEnd := FindChar(remaining[:lineEnd], ' ')
					if uriEnd == -1 {
						currState = STATE_BAD_REQUEST
						continue
					}

					queryStart := FindChar(remaining[:lineEnd], '?')
					if queryStart != -1 {
						r.URL.Path = remaining[:queryStart]
						r.URL.Query = remaining[queryStart+1 : uriEnd]
					} else {
						r.URL.Path = remaining[:uriEnd]
						r.URL.Query = ""
					}

					const httpVersionPrefix = "HTTP/"
					httpVersion := remaining[uriEnd+1 : lineEnd]
					if httpVersion[:len(httpVersionPrefix)] != httpVersionPrefix {
						currState = STATE_BAD_REQUEST
						continue
					}
					r.Version = httpVersion[len(httpVersionPrefix):]
					reqB.Consume(len(r.URL.Path) + len(r.URL.Query) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n"))
					currState = STATE_UNKNOWN
				case STATE_HEADER:
					remaining := reqB.UnconsumedString()
					lineEnd := FindChar(remaining, '\r')
					if lineEnd == -1 {
						prevState = STATE_HEADER
						currState = STATE_NEED_MORE
						continue
					}
					header := remaining[:lineEnd]
					reqB.Consume(len(header) + len("\r\n"))
					currState = STATE_UNKNOWN

				case STATE_BAD_REQUEST:
					if pipelining && respB.UnconsumedLen() > 0 {
						currState = STATE_NEED_FLUSH
						continue
					}
					WriteFull(c, []byte(ResponseBadRequest))
					Close(c)
					break connectionFor
				case STATE_METHOD_NOT_ALLOWED:
					if pipelining && respB.UnconsumedLen() > 0 {
						currState = STATE_NEED_FLUSH
						continue
					}
					WriteFull(c, []byte(ResponseMethodNotAllowed))
					break connectionFor
				case STATE_REQUEST_TIMEOUT:
					if pipelining && respB.UnconsumedLen() > 0 {
						currState = STATE_NEED_FLUSH
						continue
					}
					WriteFull(c, []byte(ResponseRequestTimeout))
					break connectionFor
				case STATE_REQUEST_ENTITY_TOO_LARGE:
					if pipelining && respB.UnconsumedLen() > 0 {
						currState = STATE_NEED_FLUSH
						continue
					}
					WriteFull(c, []byte(ResponseRequestEntityTooLarge))
					break connectionFor

				default:
					panic("invalid state")
				}
			}

			wp := (*Response)(Noescape(unsafe.Pointer(&w)))
			rp := (*Request)(Noescape(unsafe.Pointer(&r)))

			// runtime.Breakpoint()

			router(wp, rp)
			switch w.Code {
			case StatusOK:
				const statusLine = "HTTP/1.1 200 OK\r\n"
				var headers string
				switch w.ContentType {
				case "", "text/html":
					headers = "\r\nContent-Type: text/html\r\n\r\n"
				case "image/jpg":
					headers = "\r\nContent-Type: image/jpg\r\nCache-Control: max-age=604800\r\n\r\n"
				case "application/rss+xml":
					headers = "\r\nContent-Type: application/rss+xml\r\n\r\n"
				case "image/png":
					headers = "\r\nContent-Type: image/png\r\nCache-Control: max-age=604800\r\n\r\n"
				default:
					panic("unknown Content-Type '" + w.ContentType + "'")
				}

				nlength := SlicePutPositiveInt(contentLengthBuf[len(HeaderContentLength):], len(w.Body))
				contentLengthHeader := contentLengthBuf[:len(HeaderContentLength)+nlength]

				offset := len(statusLine) + len(contentLengthHeader) + len(headers)
				if offset+len(w.Body) > respB.RemainingSpace() {
					/* NOTE(anton2920): at this point w.Body is no longer a slice of respB.Buf. */
					println("Sizes:", respB.RemainingSpace(), offset+len(w.Body))
					println(respB.Start, respB.Pos, respB.End)
					panic("increase response buffer size")
				}

				remaining := respB.RemainingSlice()
				if !pipelining {
					remaining = remaining[maxResponseOffset-offset : maxResponseOffset+len(w.Body)]
				}
				copy(remaining, []byte(statusLine))
				copy(remaining[len(statusLine):], contentLengthHeader)
				copy(remaining[len(statusLine)+len(contentLengthHeader):], headers)
				if pipelining {
					copy(remaining[offset:], w.Body)
					respB.Produce(offset + len(w.Body))
				} else {
					respB.Consume(maxResponseOffset - offset)
					respB.Produce(maxResponseOffset + len(w.Body))
				}
			case StatusBadRequest:
				WriteFull(c, unsafe.Slice(unsafe.StringData(ResponseBadRequest), len(ResponseBadRequest)))
				break connectionFor
			case StatusNotFound:
				WriteFull(c, unsafe.Slice(unsafe.StringData(ResponseNotFound), len(ResponseNotFound)))
				break connectionFor
			}

			if !pipelining {
				n := respB.FullWriteTo(c)
				if n <= 0 {
					switch -n {
					case 0:
						/* End of file. */
					case EPIPE:
						/* Broken pipe. */
					case ECONNRESET:
						/* Connection reset by peer */
					default:
						println("ERROR: failed to write buffer: ", n)
					}
					break connectionFor
				}
			}
		}
		Shutdown(c, SHUT_WR)
		Close(c)
	}
}

func ListenAndServe(port uint16, router HTTPRouter) error {
	var ret, l int32

	if l = Socket(PF_INET, SOCK_STREAM, 0); l < 0 {
		return NewError("Failed to create socket: ", int(l))
	}

	var enable int32 = 1
	if ret = Setsockopt(l, SOL_SOCKET, SO_REUSEPORT, unsafe.Pointer(&enable), uint32(unsafe.Sizeof(enable))); ret != 0 {
		return NewError("Failed to set socket option to reuse port: ", int(ret))
	}

	addr := SockAddrIn{Family: AF_INET, Addr: INADDR_ANY, Port: SwapBytesInWord(port)}
	if ret = Bind(l, &addr, uint32(unsafe.Sizeof(addr))); ret < 0 {
		return NewError("Failed to bind socket to address: ", int(ret))
	}

	const backlog = 128
	if ret = Listen(l, backlog); ret < 0 {
		return NewError("Failed to listen on the socket: ", int(ret))
	}

	cc := make(chan int32)
	for i := 0; i < 8; i++ {
		go HTTPWorker(cc, router)
	}

	readTimeout := Timeval{Sec: 4}

	for {
		var c int32
		var addr SockAddrIn
		var addrLen uint32 = uint32(unsafe.Sizeof(addr))
		if c = Accept(l, &addr, &addrLen); c < 0 {
			println("ERROR: failed to accept incoming connection: ", c)
			continue
		}

		if ret = Setsockopt(c, SOL_SOCKET, SO_RCVTIMEO, unsafe.Pointer(&readTimeout), uint32(unsafe.Sizeof(readTimeout))); ret < 0 {
			return NewError("Failed to set socket read timeout: ", int(ret))
		}

		cc <- c
	}
}
