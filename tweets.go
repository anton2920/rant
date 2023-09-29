package main

import "unsafe"

var (
	TweetHTMLs [][]byte
	TweetRSSs  [][]byte
	TweetTexts [][]byte
)

func ReadTweets() {
	const tweetsPath = "tweets/"

	const tweetBeforeDate = `<div class="tweet"><div class="tweet-insides"><img class="tweet-avatar" src="/photo.jpg" alt="Profile picture"><div><div class="tweet-header"><a href="/"><b>Anton Pavlovskii</b><span>@anton2920 `
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

	TweetHTMLs = make([][]byte, 0, 1)
	TweetTexts = make([][]byte, 0, 1)

	for i := 0; ; i++ {
		tweet := make([]byte, 0, 256)

		idBufLen = SlicePutPositiveInt(unsafe.Slice(&idBuf[0], len(idBuf)), i)
		copy(unsafe.Slice(&pathBuf[len(tweetsPath)], len(pathBuf)-len(tweetsPath)), unsafe.Slice(&idBuf[0], idBufLen))

		if fd = Open(unsafe.String(&pathBuf[0], len(pathBuf)), O_RDONLY, 0); fd < 0 {
			if -fd != ENOENT {
				Fatal("Failed to open '"+string(pathBuf[:])+"': ", int(fd))
			}
			return
		}
		if ret := Fstat(fd, &st); ret < 0 {
			Fatal("Failed to get stat of '"+string(pathBuf[:])+"': ", int(ret))
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

func MonitorTweets() {
	var fd int32

	const tweetsDir = "./tweets\x00"
	if fd = Open(tweetsDir, O_RDONLY, 0); fd < 0 {
		Fatal("Failed to open '"+tweetsDir+"': ", int(fd))
	}

	tweetsKevent := Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE}
	if err := KqueueMonitor(unsafe.Slice(&tweetsKevent, 1), func(event Kevent_t) {
		println("INFO: change in tweets directory. Reloading...")
		ReadTweets()
		ConstructIndexPage()
	}); err != nil {
		FatalError(err)
	}
}
