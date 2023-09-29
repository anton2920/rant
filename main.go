package main

import "unsafe"

var (
	IndexPage    *[]byte
	TweetPage    *[]byte
	FinisherPage *[]byte
	RSSPage      *[]byte
	RSSFinisher  *[]byte
	Photo        *[]byte

	IndexPageFull []byte
	RSSPageFull   []byte
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

func IndexPageHandler(w *Response, r *Request) {
	const maxQueryLen = 256
	var queryString string

	if r.URL.Query != "" {
		if r.URL.Query[:len("Query=")] != "Query=" {
			w.Code = StatusBadRequest
			return
		}

		queryString = r.URL.Query[len("Query="):]
		if len(queryString) > maxQueryLen {
			w.Code = StatusBadRequest
			return
		}
	}

	if queryString != "" {
		var decodedQuery [maxQueryLen]byte
		decodedLen, ok := URLDecode(unsafe.Slice(&decodedQuery[0], len(decodedQuery)), queryString)
		if !ok {
			w.Code = StatusBadRequest
			return
		}

		w.Code = StatusOK
		w.Body = append(w.Body, *IndexPage...)
		for i := len(TweetHTMLs) - 1; i >= 0; i-- {
			if FindSubstring(unsafe.String(unsafe.SliceData(TweetTexts[i]), len(TweetTexts[i])), unsafe.String(&decodedQuery[0], decodedLen)) != -1 {
				w.Body = append(w.Body, TweetHTMLs[i]...)
			}
		}
		w.Body = append(w.Body, *FinisherPage...)
	} else {
		w.Code = StatusOK
		w.Body = append(w.Body, IndexPageFull...)
	}
}

func TweetPageHandler(w *Response, r *Request) {
	id, ok := StrToPositiveInt(r.URL.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(TweetHTMLs)-1) {
		w.Code = StatusNotFound
		return
	}

	w.Code = StatusOK
	w.Body = append(w.Body, *TweetPage...)
	w.Body = append(w.Body, TweetHTMLs[id]...)
	w.Body = append(w.Body, *FinisherPage...)
}

func PhotoHandler(w *Response, r *Request) {
	w.Code = StatusOK
	w.ContentType = "image/jpg"
	w.Body = append(w.Body, *Photo...)
}

func RSSPageHandler(w *Response, r *Request) {
	w.Code = StatusOK
	w.ContentType = "application/rss+xml"
	w.Body = append(w.Body, RSSPageFull...)
}

func Router(w *Response, r *Request) {
	if r.URL.Path == "/" {
		IndexPageHandler(w, r)
	} else if (len(r.URL.Path) == len("/photo.jpg")) && (r.URL.Path == "/photo.jpg") {
		PhotoHandler(w, r)
	} else if (len(r.URL.Path) == len("/favicon.ico")) && (r.URL.Path == "/favicon.ico") {
		/* Do nothing :) */
		w.Code = StatusNotFound
	} else if (len(r.URL.Path) > len("/tweet/")) && (r.URL.Path[:len("/tweet/")] == "/tweet/") {
		TweetPageHandler(w, r)
	} else if (len(r.URL.Path) == len("/rss")) && (r.URL.Path == "/rss") {
		RSSPageHandler(w, r)
	} else {
		w.Code = StatusNotFound
	}
}

func main() {
	var err error

	if IndexPage, err = ReadPage("pages/index.html"); err != nil {
		FatalError(err)
	}
	if TweetPage, err = ReadPage("pages/tweet.html"); err != nil {
		FatalError(err)
	}
	if FinisherPage, err = ReadPage("pages/finisher.html"); err != nil {
		FatalError(err)
	}
	if Photo, err = ReadPage("pages/photo.jpg"); err != nil {
		FatalError(err)
	}
	if RSSPage, err = ReadPage("pages/index.rss"); err != nil {
		FatalError(err)
	}
	if RSSFinisher, err = ReadPage("pages/finisher.rss"); err != nil {
		FatalError(err)
	}
	go MonitorPages()

	if err := ReadTweets(); err != nil {
		FatalError(err)
	}
	go MonitorTweets()

	ConstructIndexPage()
	ConstructRSSPage()

	const port = 7070
	println("Listening on 0.0.0.0:7070...")
	if err := ListenAndServe(port, Router); err != nil {
		FatalError(err)
	}
}
