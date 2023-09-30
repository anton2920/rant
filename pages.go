package main

import "unsafe"

type PageDescription struct {
	Contents **[]byte
	Filename string
}

var (
	Pages       [10][]byte
	PageKevents []Kevent_t
)

func ConstructIndexPage() {
	var totalLen int

	totalLen += len(*IndexPage)
	for _, tweet := range TweetHTMLs {
		totalLen += len(tweet)
	}

	IndexPageFull = IndexPageFull[:0]
	IndexPageFull = append(IndexPageFull, *IndexPage...)
	for i := len(TweetHTMLs) - 1; i >= 0; i-- {
		IndexPageFull = append(IndexPageFull, TweetHTMLs[i]...)
	}
	IndexPageFull = append(IndexPageFull, *FinisherPage...)
}

func ConstructRSSPage() {
	var totalLen int

	totalLen += len(*RSSPage)
	for _, tweet := range TweetRSSs {
		totalLen += len(tweet)
	}

	RSSPageFull = RSSPageFull[:0]
	RSSPageFull = append(RSSPageFull, *RSSPage...)
	for i := len(TweetRSSs) - 1; i >= 0; i-- {
		RSSPageFull = append(RSSPageFull, TweetRSSs[i]...)
	}
	RSSPageFull = append(RSSPageFull, *RSSFinisher...)
}

func MonitorPages() {
	if err := KqueueMonitor(PageKevents, func(event Kevent_t) error {
		var err error

		println("INFO: page has been changed. Reloading...")
		if Pages[event.Ident], err = ReadEntireFile(int32(event.Ident)); err != nil {
			return err
		}
		ConstructIndexPage()
		return nil
	}); err != nil {
		FatalError(err)
	}
}

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

func ReadPages(ps []PageDescription) error {
	var err error

	for _, p := range ps {
		if *p.Contents, err = ReadPage(p.Filename); err != nil {
			return err
		}
	}

	return nil
}
