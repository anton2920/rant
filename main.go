package main

import "unsafe"

var (
	IndexPage    *[]byte
	TweetPage    *[]byte
	FinisherPage *[]byte
	RSSPage      *[]byte
	RSSFinisher  *[]byte
	Photo        *[]byte
	RSSPhoto     *[]byte

	IndexPageFull []byte
	RSSPageFull   []byte
)

/* NOTE(anton2920): Noescape hides a pointer from escape analysis. Noescape is the identity function but escape analysis doesn't think the output depends on the input. Noescape is inlined and currently compiles down to zero instructions. */
//go:nosplit
func Noescape(p unsafe.Pointer) unsafe.Pointer {
	x := uintptr(p)
	return unsafe.Pointer(x ^ 0)
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

func RSSPhotoHandler(w *Response, r *Request) {
	w.Code = StatusOK
	w.ContentType = "image/png"
	w.Body = append(w.Body, *RSSPhoto...)
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
	} else if (len(r.URL.Path) == len("/rss.png")) && (r.URL.Path == "/rss.png") {
		RSSPhotoHandler(w, r)
	} else {
		w.Code = StatusNotFound
	}
}

func main() {
	if err := ReadPages([]PageDescription{
		{&IndexPage, "pages/index.html"},
		{&TweetPage, "pages/tweet.html"},
		{&FinisherPage, "pages/finisher.html"},
		{&Photo, "pages/photo.jpg"},
		{&RSSPage, "pages/index.rss"},
		{&RSSFinisher, "pages/finisher.rss"},
		{&RSSPhoto, "pages/rss.png"},
	}); err != nil {
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
