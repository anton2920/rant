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
			w.WriteBuiltinError(HTTPStatusBadRequest)
			return
		}

		queryString = r.URL.Query[len("Query="):]
		if len(queryString) > maxQueryLen {
			w.WriteBuiltinError(HTTPStatusBadRequest)
			return
		}
	}

	if queryString != "" {
		var decodedQuery [maxQueryLen]byte
		decodedLen, ok := URLDecode(unsafe.Slice(&decodedQuery[0], len(decodedQuery)), queryString)
		if !ok {
			w.WriteBuiltinError(HTTPStatusBadRequest)
			return
		}

		w.WriteResponseNoCopyFunc("text/html", func(w *HTTPResponse) {
			w.WriteBodyNoCopy(*IndexPage)
			for i := len(TweetHTMLs) - 1; i >= 0; i-- {
				if FindSubstring(unsafe.String(unsafe.SliceData(TweetTexts[i]), len(TweetTexts[i])), unsafe.String(&decodedQuery[0], decodedLen)) != -1 {
					w.WriteBodyNoCopy(TweetHTMLs[i])
				}
			}
			w.WriteBodyNoCopy(*FinisherPage)
		})
	} else {
		w.WriteResponseNoCopy("text/html", IndexPageFull)
	}
}

func PlaintextHandler(w *HTTPResponse, r *HTTPRequest) {
	const response = "Hello, world!\n"
	w.WriteResponseNoCopy("text/plain", unsafe.Slice(unsafe.StringData(response), len(response)))
}

func TweetPageHandler(w *HTTPResponse, r *HTTPRequest) {
	id, ok := StrToPositiveInt(r.URL.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(TweetHTMLs)-1) {
		w.WriteBuiltinError(HTTPStatusNotFound)
		return
	}

	w.WriteResponseNoCopyFunc("text/html", func(w *HTTPResponse) {
		w.WriteBodyNoCopy(*TweetPage)
		w.WriteBodyNoCopy(TweetHTMLs[id])
		w.WriteBodyNoCopy(*FinisherPage)
	})
}

func PhotoHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteResponseNoCopy("image/jpg\r\nCache-Control: max-age=604800", *Photo)
}

func RSSPageHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteResponseNoCopy("application/rss+xml", RSSPageFull)
}

func RSSPhotoHandler(w *HTTPResponse, r *HTTPRequest) {
	w.WriteResponseNoCopy("image/png\r\nCache-Control: max-age=604800", *RSSPhoto)
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
		w.WriteBuiltinError(HTTPStatusNotFound)
	}
}

func main() {
	IndexPage = Must(ReadPage("pages/index.html"))
	TweetPage = Must(ReadPage("pages/tweet.html"))
	FinisherPage = Must(ReadPage("pages/finisher.html"))
	Photo = Must(ReadPage("pages/photo.jpg"))
	RSSPage = Must(ReadPage("pages/index.rss"))
	RSSFinisher = Must(ReadPage("pages/finisher.rss"))
	RSSPhoto = Must(ReadPage("pages/rss.png"))
	go MonitorPages()

	if err := ReadTweets(); err != nil {
		FatalError("Failed to read tweets: ", err)
	}
	go MonitorTweets()

	ConstructIndexPage()
	ConstructRSSPage()

	const port = 7070
	println("Listening on 0.0.0.0:7070...")
	if err := ListenAndServe(port, Router); err != nil {
		FatalError("Failed to start HTTP server: ", err)
	}
}
