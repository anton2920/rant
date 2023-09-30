package main

import "unsafe"

var (
	TweetHTMLs [][]byte
	TweetRSSs  [][]byte
	TweetTexts [][]byte
)

func ReadTweets() error {
	const tweetsPath = "tweets/"

	const tweetHTMLBeforeDate = `<div class="tweet"><div class="tweet-insides"><img class="tweet-avatar" src="/photo.jpg" alt="Profile picture"><div><div class="tweet-header"><a href="/"><b>Anton Pavlovskii</b><span>@anton2920 `
	const tweetHTMLBeforeID = `</span></a></div><a href="/tweet/`
	const tweetHTMLBeforeText = `"><p>`
	const tweetHTMLAfterText = `</p></a></div></div></div>`

	const tweetRSSBeforeTitle = `<item><title>Tweet #`
	const tweetRSSBeforeDesc = `</title><description>`
	const tweetRSSBeforeLink = `</description><link>https://rant.anton2920.ru/tweet/`
	const tweetRSSBeforeDate = `</link><pubDate>`
	const tweetRSSAfterDate = `</pubDate></item>`

	var err error

	var pathBuf [PATH_MAX]byte

	var idBuf [10]byte
	var idBufLen int

	var dateBuf [50]byte
	var dateBufLen int

	var text []byte

	var fd int32
	var st Stat

	copy(unsafe.Slice(&pathBuf[0], len(pathBuf)), []byte(tweetsPath))

	TweetTexts = TweetTexts[:0]
	TweetHTMLs = TweetHTMLs[:0]
	TweetRSSs = TweetRSSs[:0]

	for i := 0; ; i++ {
		idBufLen = SlicePutPositiveInt(unsafe.Slice(&idBuf[0], len(idBuf)), i)
		copy(unsafe.Slice(&pathBuf[len(tweetsPath)], len(pathBuf)-len(tweetsPath)), unsafe.Slice(&idBuf[0], idBufLen))

		if fd = Open(unsafe.String(&pathBuf[0], len(pathBuf)), O_RDONLY, 0); fd < 0 {
			if -fd != ENOENT {
				return NewError("Failed to open '"+string(pathBuf[:])+"': ", int(fd))
			}
			return nil
		}
		if ret := Fstat(fd, &st); ret < 0 {
			return NewError("Failed to get stat of '"+string(pathBuf[:])+"': ", int(ret))
		}
		tm := TimeToTm(st.Birthtime.Sec)
		dateBufLen = SlicePutTm(unsafe.Slice(&dateBuf[0], len(dateBuf)), tm)

		if text, err = ReadEntireFile(fd); err != nil {
			return err
		}
		TweetTexts = append(TweetTexts, text)
		Close(fd)

		tweet := make([]byte, 0, 4*1024)
		tweet = append(tweet, tweetHTMLBeforeDate...)
		tweet = append(tweet, unsafe.Slice(&dateBuf[0], dateBufLen)...)
		tweet = append(tweet, tweetHTMLBeforeID...)
		tweet = append(tweet, unsafe.Slice(&idBuf[0], idBufLen)...)
		tweet = append(tweet, tweetHTMLBeforeText...)
		tweet = append(tweet, text...)
		tweet = append(tweet, tweetHTMLAfterText...)
		TweetHTMLs = append(TweetHTMLs, tweet)

		dateBufLen = SlicePutTmRFC822(unsafe.Slice(&dateBuf[0], len(dateBuf)), tm)
		tweet = make([]byte, 0, 4*1024)
		tweet = tweet[:0]
		tweet = append(tweet, tweetRSSBeforeTitle...)
		tweet = append(tweet, unsafe.Slice(&idBuf[0], idBufLen)...)
		tweet = append(tweet, tweetRSSBeforeDesc...)
		tweet = append(tweet, text...)
		tweet = append(tweet, tweetRSSBeforeLink...)
		tweet = append(tweet, unsafe.Slice(&idBuf[0], idBufLen)...)
		tweet = append(tweet, tweetRSSBeforeDate...)
		tweet = append(tweet, unsafe.Slice(&dateBuf[0], dateBufLen)...)
		tweet = append(tweet, tweetRSSAfterDate...)
		TweetRSSs = append(TweetRSSs, tweet)
	}
}

func MonitorTweets() {
	var fd int32

	const tweetsDir = "./tweets\x00"
	if fd = Open(tweetsDir, O_RDONLY, 0); fd < 0 {
		Fatal("Failed to open '"+tweetsDir+"': ", int(fd))
	}

	tweetsKevent := Kevent_t{Ident: uintptr(fd), Filter: EVFILT_VNODE, Flags: EV_ADD | EV_CLEAR, Fflags: NOTE_WRITE}
	if err := KqueueMonitor(unsafe.Slice(&tweetsKevent, 1), func(event Kevent_t) error {
		println("INFO: change in tweets directory. Reloading...")
		if err := ReadTweets(); err != nil {
			return err
		}
		ConstructIndexPage()
		ConstructRSSPage()
		return nil
	}); err != nil {
		FatalError(err)
	}
}
