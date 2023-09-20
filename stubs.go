package main

//go:nosplit
func Exit(status int)

//go:nosplit
func Socket(domain, typ, protocol int) int

//go:nosplit
func Write(fd int, buf string) int
