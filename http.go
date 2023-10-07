package main

import (
	"runtime"
	"unsafe"
)

type HTTPRequestParser struct {
	State int
}

type HTTPRequest struct {
	Method  string
	URL     URL
	Version string
}

type HTTPResponse struct {
	Code        int
	ContentType string
	Body        []byte
}

type HTTPContext struct {
	RequestBuffer  CircularBuffer
	ResponseBuffer CircularBuffer

	PendingRequests  SyncQueue[HTTPRequest]
	PendingResponses SyncQueue[HTTPResponse]

	InProgressRequest *HTTPRequest
	Parser            HTTPRequestParser

	/* Check must be the same as pointer's bit, if context is in use. */
	Check uintptr
}

type HTTPRouter func(w *HTTPResponse, r *HTTPRequest)

const (
	HTTPStatusOK                    = 200
	HTTPStatusBadRequest            = 400
	HTTPStatusNotFound              = 404
	HTTPStatusMethodNotAllowed      = 405
	HTTPStatusRequestTimeout        = 408
	HTTPStatusRequestEntityTooLarge = 413
)

const (
	HTTPResponseBadRequest            = "HTTP/1.1 400 Bad HTTPRequest\r\nContent-Type: text/html\r\nContent-Length: 175\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>400 Bad HTTPRequest</title></head><body><h1>400 Bad HTTPRequest</h1><p>Your browser sent a request that this server could not understand.</p></body></html>"
	HTTPResponseNotFound              = "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nContent-Length: 152\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p>The requested URL was not found on this server.</p></body></html>"
	HTTPResponseMethodNotAllowed      = "HTTP/1.1 405 Method Not Allowed\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	HTTPResponseRequestTimeout        = "HTTP/1.1 408 HTTPRequest Timeout\r\nContent-Type: text/html\r\nContent-Length: ...\r\nConnection: close\r\n\r\n"
	HTTPResponseRequestEntityTooLarge = "HTTP/1.1 413 HTTPRequest Entity Too Large\r\nContent-Type: text/html\r\nConent-Length: ...\r\nConnection: close\r\n\r\n"
)

const (
	HTTPHeaderContentLengthPrefix = `Content-Length: `
	HTTPHeaderDatePrefix          = `Date: `
	HTTPHeaderHost                = `Host: rant`
)

const (
	HTTP_STATE_UNKNOWN = iota

	HTTP_STATE_METHOD
	HTTP_STATE_URI
	HTTP_STATE_VERSION
	HTTP_STATE_HEADER

	HTTP_STATE_DONE
)

const (
	HTTP_EVENT_DECODE = iota
	HTTP_EVENT_EXECUTE
	HTTP_EVENT_WRITEBACK
)

func NewHTTPContext() *HTTPContext {
	var err error

	c := new(HTTPContext)
	c.Parser.State = HTTP_STATE_METHOD

	if c.RequestBuffer, err = NewCircularBuffer(PageSize); err != nil {
		return nil
	}
	if c.ResponseBuffer, err = NewCircularBuffer(128 * PageSize); err != nil {
		return nil
	}

	return c
}

func NewHTTPRequest() *HTTPRequest {
	return new(HTTPRequest)
}

func NewHTTPResponse() *HTTPResponse {
	w := new(HTTPResponse)
	w.Body = make([]byte, 16*1024)
	return w
}

func GetContextAndCheck(ptr unsafe.Pointer) (*HTTPContext, uintptr) {
	uptr := uintptr(ptr)

	check := uptr & 0x1
	ctx := (*HTTPContext)(unsafe.Pointer(uptr - check))

	return ctx, check
}

func HTTPWorker(kq int32, l int32, ctxPool *SyncPool[HTTPContext], rPool *SyncPool[HTTPRequest], wPool *SyncPool[HTTPResponse], router HTTPRouter) {
	var contentLengthBuf []byte
	contentLengthBuf = make([]byte, len(HTTPHeaderContentLengthPrefix)+10)
	copy(contentLengthBuf, []byte(HTTPHeaderContentLengthPrefix))

	var pinner runtime.Pinner
	var events [256]Kevent_t
	var nevents, i int32

	runtime.LockOSThread()

	/* NOTE(anton2920): this is indended to be optimized for high throughput.
	 * Pipeline consists of following steps:
	 * 0. Accept: accept new connection.
	 * 1. Fetch: read bytes from socket.
	 * 2. Decode: parse bytes and fill HTTPRequest struct.
	 * 3. Execute: do router(), which fills HTTPResponse struct.
	 * 4. Writeback: write responses to buffer.
	 * 5. Retire: flush buffer to socket.
	 */
	for {
		if nevents = Kevent(kq, nil, unsafe.Slice(&events[0], len(events)), nil); nevents < 0 {
			if -nevents == EINTR {
				continue
			}
			println("ERROR: failed to get requested kernel events: ", nevents)
		}
		for i = 0; i < nevents; i++ {
			e := events[i]
			c := int32(e.Ident)

			// println("EVENT", e.Ident, e.Filter, e.Fflags&0xF, e.Data)

			if c == l {
				if c = Accept(l, nil, nil); c < 0 {
					println("ERROR: failed to accept new connection: ", c)
					continue
				}

				ctx := SyncPoolGet(ctxPool)
				if ctx == nil {
					Fatal("Failed to acquire new HTTP context", 0)
				}
				pinner.Pin(ctx)

				udata := unsafe.Pointer(uintptr(unsafe.Pointer(ctx)) | ctx.Check)
				events := [...]Kevent_t{
					{Ident: uintptr(c), Filter: EVFILT_READ, Flags: EV_ADD | EV_CLEAR, Udata: udata},
					{Ident: uintptr(c), Filter: EVFILT_USER, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_FFCOPY},
					{Ident: uintptr(c), Filter: EVFILT_WRITE, Flags: EV_ADD | EV_CLEAR | EV_DISABLE | EV_DISPATCH, Udata: udata},
				}
				if ret := Kevent(kq, unsafe.Slice(&events[0], len(events)), nil, nil); ret < 0 {
					println("ERROR: failed to add new events to kqueue", ret)
					SyncPoolPut(ctxPool, ctx)
					continue
				}
				continue
			}

			ctx, check := GetContextAndCheck(e.Udata)
			if check != ctx.Check {
				// println("Invalid event", ctx, check, ctx.Check)
				continue
			}

			if (e.Flags & EV_EOF) != 0 {
				// println("Client disconnected: ", e.Ident)
				ctx.Check = 1 - check
				SyncPoolPut(ctxPool, ctx)
				Close(c)
				continue
			}

			switch e.Filter {
			case EVFILT_READ:
				rBuf := &ctx.RequestBuffer
				if rBuf.RemainingSpace() == 0 {
					if ctx.InProgressRequest != nil {
						w := SyncPoolGet(wPool)
						w.Code = HTTPStatusRequestTimeout
						SyncQueuePut(&ctx.PendingResponses, w)
					}
					e.Flags |= EV_DISABLE
					Kevent(kq, unsafe.Slice(&e, 1), nil, nil)
					continue
				}
				rBuf.ReadFrom(c)

				e.Flags = 0
				e.Filter = EVFILT_USER
				e.Fflags = NOTE_TRIGGER | NOTE_FFCOPY | HTTP_EVENT_DECODE
				e.Data = 0
				if ret := Kevent(kq, unsafe.Slice(&e, 1), nil, nil); ret < 0 {
					println("ERROR: failed to add trigger event: ", ret)
				}

			case EVFILT_USER:
				switch e.Fflags & 0xF {
				default:
					println(e.Fflags)
					panic("invalid fflags for user event")
				case HTTP_EVENT_DECODE:
					rBuf := &ctx.RequestBuffer
					for {
						if ctx.InProgressRequest == nil {
							ctx.InProgressRequest = SyncPoolGet(rPool)
						}
						r := ctx.InProgressRequest

					parserLoop:
						for ctx.Parser.State != HTTP_STATE_DONE {
							switch ctx.Parser.State {
							default:
								panic(ctx.Parser.State)
							case HTTP_STATE_UNKNOWN:
								unconsumed := rBuf.UnconsumedString()
								if len(unconsumed) < 2 {
									break parserLoop
								}
								if unconsumed[:2] == "\r\n" {
									rBuf.Consume(len("\r\n"))
									ctx.Parser.State = HTTP_STATE_DONE
								} else {
									ctx.Parser.State = HTTP_STATE_HEADER
								}

							case HTTP_STATE_METHOD:
								unconsumed := rBuf.UnconsumedString()
								if len(unconsumed) < 3 {
									break parserLoop
								}
								switch unconsumed[:3] {
								case "GET":
									r.Method = "GET"
								default:
									w := SyncPoolGet(wPool)
									w.Code = HTTPStatusMethodNotAllowed
									SyncQueuePut(&ctx.PendingResponses, w)
								}
								rBuf.Consume(len(r.Method) + 1)
								ctx.Parser.State = HTTP_STATE_URI
							case HTTP_STATE_URI:
								unconsumed := rBuf.UnconsumedString()
								lineEnd := FindChar(unconsumed, '\r')
								if lineEnd == -1 {
									break parserLoop
								}

								uriEnd := FindChar(unconsumed[:lineEnd], ' ')
								if uriEnd == -1 {
									w := SyncPoolGet(wPool)
									w.Code = HTTPStatusBadRequest
									SyncQueuePut(&ctx.PendingResponses, w)
								}

								queryStart := FindChar(unconsumed[:lineEnd], '?')
								if queryStart != -1 {
									r.URL.Path = unconsumed[:queryStart]
									r.URL.Query = unconsumed[queryStart+1 : uriEnd]
								} else {
									r.URL.Path = unconsumed[:uriEnd]
									r.URL.Query = ""
								}

								const httpVersionPrefix = "HTTP/"
								httpVersion := unconsumed[uriEnd+1 : lineEnd]
								if httpVersion[:len(httpVersionPrefix)] != httpVersionPrefix {
									w := SyncPoolGet(wPool)
									w.Code = HTTPStatusBadRequest
									SyncQueuePut(&ctx.PendingResponses, w)
								}
								r.Version = httpVersion[len(httpVersionPrefix):]
								rBuf.Consume(len(r.URL.Path) + len(r.URL.Query) + 1 + len(httpVersionPrefix) + len(r.Version) + len("\r\n"))
								ctx.Parser.State = HTTP_STATE_UNKNOWN
							case HTTP_STATE_HEADER:
								unconsumed := rBuf.UnconsumedString()
								lineEnd := FindChar(unconsumed, '\r')
								if lineEnd == -1 {
									break parserLoop
								}
								header := unconsumed[:lineEnd]
								rBuf.Consume(len(header) + len("\r\n"))
								ctx.Parser.State = HTTP_STATE_UNKNOWN
							}
						}

						if ctx.Parser.State != HTTP_STATE_DONE {
							if ctx.Parser.State == HTTP_STATE_METHOD {
								SyncPoolPut(rPool, r)
								ctx.InProgressRequest = nil
							}
							break
						}
						SyncQueuePut(&ctx.PendingRequests, r)
						ctx.InProgressRequest = nil
						ctx.Parser.State = HTTP_STATE_METHOD
					}

					e.Flags = 0
					e.Filter = EVFILT_USER
					e.Fflags = NOTE_TRIGGER | NOTE_FFCOPY | HTTP_EVENT_EXECUTE
					e.Data = 0
					if ret := Kevent(kq, unsafe.Slice(&e, 1), nil, nil); ret < 0 {
						println("ERROR: failed to add trigger event: ", ret)
					}
				case HTTP_EVENT_EXECUTE:
					for {
						r := SyncQueueGet(&ctx.PendingRequests)
						if r == nil {
							break
						}
						w := SyncPoolGet(wPool)
						w.Body = w.Body[:0]
						w.ContentType = ""

						router(w, r)

						SyncPoolPut(rPool, r)
						SyncQueuePut(&ctx.PendingResponses, w)
					}

					e.Flags = 0
					e.Filter = EVFILT_USER
					e.Fflags = NOTE_TRIGGER | NOTE_FFCOPY | HTTP_EVENT_WRITEBACK
					e.Data = 0
					if ret := Kevent(kq, unsafe.Slice(&e, 1), nil, nil); ret < 0 {
						println("ERROR: failed to add trigger event: ", ret)
					}
				case HTTP_EVENT_WRITEBACK:
					wBuf := &ctx.ResponseBuffer

					for {
						w := SyncQueueGet(&ctx.PendingResponses)
						if w == nil {
							break
						}

						remaining := wBuf.RemainingSlice()
						switch w.Code {
						case HTTPStatusOK:
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

							nlength := SlicePutPositiveInt(contentLengthBuf[len(HTTPHeaderContentLengthPrefix):], len(w.Body))
							contentLengthHeader := contentLengthBuf[:len(HTTPHeaderContentLengthPrefix)+nlength]

							offset := len(statusLine) + len(contentLengthHeader) + len(headers)
							if offset+len(w.Body) > wBuf.RemainingSpace() {
								println("Sizes:", wBuf.RemainingSpace(), offset+len(w.Body))
								println(wBuf.Head, wBuf.Tail)
								panic("increase response buffer size")
							}

							copy(remaining, []byte(statusLine))
							copy(remaining[len(statusLine):], contentLengthHeader)
							copy(remaining[len(statusLine)+len(contentLengthHeader):], headers)
							copy(remaining[offset:], w.Body)
							wBuf.Produce(offset + len(w.Body))
						case HTTPStatusBadRequest:
							copy(remaining, unsafe.Slice(unsafe.StringData(HTTPResponseBadRequest), len(HTTPResponseBadRequest)))
							wBuf.Produce(len(HTTPResponseBadRequest))
						case HTTPStatusNotFound:
							copy(remaining, unsafe.Slice(unsafe.StringData(HTTPResponseNotFound), len(HTTPResponseNotFound)))
							wBuf.Produce(len(HTTPResponseNotFound))
						}

						SyncPoolPut(wPool, w)
					}

					e.Flags = EV_ENABLE
					e.Filter = EVFILT_WRITE
					e.Fflags = 0
					if ret := Kevent(kq, unsafe.Slice(&e, 1), nil, nil); ret < 0 {
						println("ERROR: failed to add trigger event: ", ret)
					}
				}
			case EVFILT_WRITE:
				ctx.ResponseBuffer.WriteTo(c)
			}
		}
	}
}

func ListenAndServe(port uint16, router HTTPRouter) error {
	var ret, l, kq int32

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

	if kq = Kqueue(); kq < 0 {
		return NewError("Failed to open kernel queue: ", int(kq))
	}
	event := Kevent_t{Ident: uintptr(l), Filter: EVFILT_READ, Flags: EV_ADD | EV_CLEAR}
	if ret = Kevent(kq, unsafe.Slice(&event, 1), nil, nil); ret < 0 {
		return NewError("Failed to add event for listener socket: ", int(ret))
	}

	ctxPool := NewSyncPool[HTTPContext](16*1024, NewHTTPContext)
	rPool := NewSyncPool[HTTPRequest](16*1024, NewHTTPRequest)
	wPool := NewSyncPool[HTTPResponse](16*1024, NewHTTPResponse)

	const nworkers = 4
	for i := 0; i < nworkers-1; i++ {
		go HTTPWorker(kq, l, ctxPool, rPool, wPool, router)
	}
	HTTPWorker(kq, l, ctxPool, rPool, wPool, router)

	return nil
}
