package main

import "unsafe"

const (
	PageSize     = 4 * 1024
	HugePageSize = 2 * 1024 * 1024
)

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

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) {
	const maxQueryLen = 256
	var queryString string

	if r.URL.Query != "" {
		if r.URL.Query[:len("Query=")] != "Query=" {
			w.WriteBuiltinResponse(HTTPStatusBadRequest)
			return
		}

		queryString = r.URL.Query[len("Query="):]
		if len(queryString) > maxQueryLen {
			w.WriteBuiltinResponse(HTTPStatusBadRequest)
			return
		}
	}

	if queryString != "" {
		var decodedQuery [maxQueryLen]byte
		decodedLen, ok := URLDecode(unsafe.Slice(&decodedQuery[0], len(decodedQuery)), queryString)
		if !ok {
			w.WriteBuiltinResponse(HTTPStatusBadRequest)
			return
		}

		w.StartResponse(HTTPStatusOK, "text/html")
		w.WriteResponseBody(*IndexPage)
		for i := len(TweetHTMLs) - 1; i >= 0; i-- {
			if FindSubstring(unsafe.String(unsafe.SliceData(TweetTexts[i]), len(TweetTexts[i])), unsafe.String(&decodedQuery[0], decodedLen)) != -1 {
				w.WriteResponseBody(TweetHTMLs[i])
			}
		}
		w.WriteResponseBody(*FinisherPage)
		w.FinishResponse()
	} else {
		w.WriteCompleteResponse(HTTPStatusOK, "text/html", IndexPageFull)
	}
}

func PlaintextHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteCompleteResponse(HTTPStatusOK, "text/plain", []byte("Hello, world!\n"))
}

func TweetPageHandler(w *HTTPResponse, r *HTTPRequest) {
	id, ok := StrToPositiveInt(r.URL.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(TweetHTMLs)-1) {
		w.WriteBuiltinResponse(HTTPStatusNotFound)
		return
	}

	w.StartResponseWithSize(HTTPStatusOK, "text/html", len(*TweetPage)+len(TweetHTMLs[id])+len(*FinisherPage))
	w.WriteResponseBody(*TweetPage)
	w.WriteResponseBody(TweetHTMLs[id])
	w.WriteResponseBody(*FinisherPage)
}

func PhotoHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteCompleteResponse(HTTPStatusOK, "image/jpg\r\nCache-Control: max-age=604800", *Photo)
}

func RSSPageHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteCompleteResponse(HTTPStatusOK, "application/rss+xml", RSSPageFull)
}

func RSSPhotoHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteCompleteResponse(HTTPStatusOK, "image/png\r\nCache-Control: max-age=604800", *RSSPhoto)
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	if r.URL.Path == "/" {
		IndexPageHandler(w, r)
	} else if r.URL.Path == "/plaintext" {
		PlaintextHandler(w, r)
	} else if r.URL.Path == "/photo.jpg" {
		PhotoHandler(w, r)
	} else if (len(r.URL.Path) > len("/tweet/")) && (r.URL.Path[:len("/tweet/")] == "/tweet/") {
		TweetPageHandler(w, r)
	} else if r.URL.Path == "/rss" {
		RSSPageHandler(w, r)
	} else if r.URL.Path == "/rss.png" {
		RSSPhotoHandler(w, r)
	} else {
		w.WriteBuiltinResponse(HTTPStatusNotFound)
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
