package main

var (
	Pages       [15][]byte
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
		FatalError("Failed to monitor pages:", err)
	}
}

func ReadPage(name string) (*[]byte, error) {
	var err error

	fd, err := Open(name, O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	PageKevents = append(PageKevents, Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE})

	Pages[fd], err = ReadEntireFile(fd)
	if err != nil {
		return nil, err
	}
	return &Pages[fd], nil
}

func Must[T any](ret T, err error) T {
	if err != nil {
		FatalError("Actiion must succeed, but it failed:", err)
	}
	return ret
}
