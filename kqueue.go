package main

import "unsafe"

type Kevent_t struct {
	Ident  uintptr
	Filter int16
	Flags  uint16
	Fflags uint32
	Data   int
	Udata  unsafe.Pointer
	Ext    [4]uint
}

type KqueueCb func(Kevent_t) error

const (
	/* From <sys/event.h>. */
	EVFILT_READ  = -1
	EVFILT_WRITE = -2
	EVFILT_VNODE = -4
	EVFILT_USER  = -11

	EV_ADD      = 0x0001
	EV_ENABLE   = 0x0004
	EV_DISABLE  = 0x0008
	EV_ONESHOT  = 0x0010
	EV_CLEAR    = 0x0020
	EV_DISPATCH = 0x0080

	EV_EOF = 0x8000

	NOTE_FFCOPY  = 0xc0000000
	NOTE_TRIGGER = 0x01000000

	NOTE_WRITE = 0x0002
)

func KqueueMonitor(eventlist []Kevent_t, cb KqueueCb) error {
	var kq, nevents int32

	if kq = Kqueue(); kq < 0 {
		return NewError("Failed to open a kernel queue: ", int(kq))
	}

	if nevents = Kevent(kq, eventlist, nil, nil); nevents < 0 {
		return NewError("Failed to register kernel events: ", int(nevents))
	}

	var event Kevent_t
	for {
		if nevents = Kevent(kq, nil, unsafe.Slice(&event, 1), nil); nevents < 0 {
			if -nevents != EINTR {
				return NewError("Failed to get kernel events: ", int(nevents))
			}
			continue
		} else if nevents > 0 {
			if err := cb(event); err != nil {
				return err
			}
		}

		/* NOTE(anton2920): sleep to prevent runaway events. */
		SleepFull(Timespec{Nsec: 200000000})
	}
}
