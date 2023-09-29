package main

import "unsafe"

var (
	Pages       [10][]byte
	PageKevents []Kevent_t
)

func ReadPage(name string) (*[]byte, error) {
	var err error

	var nameBuf [2 * PATH_MAX]byte
	var fd int32

	/* NOTE(anton2920): this sh**t is needed, because open(2) requires '\0'-terminated string. */
	for i := 0; i < len(name); i++ {
		nameBuf[i] = name[i]
	}
	if fd = Open(unsafe.String(&nameBuf[0], len(name)+1), O_RDONLY, 0); fd < 0 {
		return nil, NewError("Failed to open '"+name+"': ", int(fd))
	}
	PageKevents = append(PageKevents, Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE})

	Pages[fd], err = ReadEntireFile(fd)
	if err != nil {
		return nil, err
	}
	return &Pages[fd], nil
}

func MonitorPages() {
	if err := KqueueMonitor(PageKevents, func(event Kevent_t) error {
		var err error

		println("INFO: page has been changed. Reloading...")
		if Pages[event.Ident], err = ReadEntireFile(int32(event.Ident)); err != nil {
			return err
		}
		return nil
	}); err != nil {
		FatalError(err)
	}
}
