package main

import (
	"runtime"
	"unsafe"
)

type Request struct {
	URL URL
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
	ResponseBadRequest = "HTTP/1.1 400 Bad Request\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>400 Bad Request</title></head><body><h1>400 Bad Request</h1><p>Your browser sent a request that this server could not understand.</p></body></html>"
	ResponseNotFound   = "HTTP/1.1 404 Not Found\r\nContent-Type: text/html\r\nConnection: close\r\n\r\n<!DOCTYPE html><head><title>404 Not Found</title></head><body><h1>404 Not Found</h1><p>The requested URL was not found on this server.</p></body></html>"
)

func HTTPWorker(cc <-chan int32, router HTTPRouter) {
	var buffer [512]byte

	var w Response
	w.Body = make([]byte, 0, 4*1024)

	for c := range cc {
		var r Request

		Read(c, unsafe.Slice(&buffer[0], len(buffer)))
		if unsafe.String(&buffer[0], 3) == "GET" {
			lineEnd := FindChar(unsafe.String(&buffer[4], len(buffer)-4), '\r')
			requestLine := unsafe.String(&buffer[4], lineEnd-1) /* without method. */

			pathEnd := FindChar(requestLine, '?')
			if pathEnd != -1 {
				/* With query. */
				r.URL.Path = unsafe.String(unsafe.StringData(requestLine), pathEnd)

				queryStart := pathEnd + 1
				queryEnd := FindChar(unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(unsafe.StringData(requestLine)), queryStart)), len(requestLine)-queryStart), ' ')
				r.URL.Query = unsafe.String((*byte)(unsafe.Add(unsafe.Pointer(unsafe.StringData(requestLine)), queryStart)), queryEnd)
			} else {
				/* No query. */
				pathEnd = FindChar(requestLine, ' ')
				r.URL.Path = unsafe.String(unsafe.StringData(requestLine), pathEnd)
			}
		} else {
			WriteFull(c, []byte(ResponseBadRequest))
			Close(c)
			clear(unsafe.Slice(&buffer[0], len(buffer)))
			continue
		}

		router(&w, &r)
		switch w.Code {
		case StatusOK:
			WriteFull(c, []byte("HTTP/1.1 200 OK\r\nConnection: close\r\n"))
			switch w.ContentType {
			case "", "text/html":
				WriteFull(c, []byte("Content-Type: text/html\r\n\r\n"))
			case "image/jpg":
				WriteFull(c, []byte("Content-Type: image/jpg\r\nCache-Control: max-age=604800\r\n\r\n"))
			case "application/rss+xml":
				WriteFull(c, []byte("Content-Type: application/rss+xml\r\n\r\n"))
			default:
				panic("unknown Content-Type")
			}
			WriteFull(c, w.Body)
		case StatusBadRequest:
			WriteFull(c, []byte(ResponseBadRequest))
		case StatusNotFound:
			WriteFull(c, []byte(ResponseNotFound))
		}

		Shutdown(c, SHUT_WR)
		Close(c)

		clear(unsafe.Slice(&buffer[0], len(buffer)))
		w.Body = w.Body[:0]
		w.ContentType = ""
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
	if ret = Listen(l, backlog); ret != 0 {
		return NewError("Failed to listen on the socket: ", int(ret))
	}

	cc := make(chan int32)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go HTTPWorker(cc, router)
	}

	for {
		var c int32
		var addr SockAddrIn
		var addrLen uint32 = uint32(unsafe.Sizeof(addr))
		if c = Accept(l, &addr, &addrLen); c < 0 {
			continue
		}

		cc <- c
	}
}
