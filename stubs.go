package main

import "unsafe"

//go:noescape
//go:nosplit
func Accept(s int32, addr *SockAddr, addrlen *uint32) int32

//go:noescape
//go:nosplit
func Bind(s int32, addr *SockAddr, addrlen uint32) int32

//go:nosplit
func Close(fd int32) int32

//go:nosplit
func Exit(status int32)

//go:nosplit
func Listen(s int32, backlog int32) int32

//go:noescape
//go:nosplit
func Setsockopt(s, level, optname int32, optval unsafe.Pointer, optlen uint32) int32

//go:nosplit
func Socket(domain, typ, protocol int32) int32

//go:noescape
//go:nosplit
func Write(fd int32, buf []byte) int
