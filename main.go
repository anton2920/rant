package main

import "unsafe"

const PageFinisher = "</div></div></div></body></html>"

var (
	IndexPage *[]byte
	TweetPage *[]byte
	RSSPage   *[]byte
	Photo     *[]byte

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
}

func IndexPageHandler(w *HTTPResponse, r *HTTPRequest) {
	const maxQueryLen = 1024
	var queryString string

	if r.URL.Query != "" {
		if r.URL.Query[:len("Query=")] != "Query=" {
			w.Code = HTTPStatusBadRequest
			return
		}

		queryString = r.URL.Query[len("Query="):]
		if len(queryString) > maxQueryLen {
			w.Code = HTTPStatusBadRequest
			return
		}
	}

	if queryString != "" {
		var decodedQuery [maxQueryLen]byte
		decodedLen, ok := URLDecode(unsafe.Slice(&decodedQuery[0], len(decodedQuery)), queryString)
		if !ok {
			w.Code = HTTPStatusBadRequest
			return
		}

		w.Code = HTTPStatusOK
		w.Body = append(w.Body, *IndexPage...)
		for i := len(TweetHTMLs) - 1; i >= 0; i-- {
			if FindSubstring(unsafe.String(unsafe.SliceData(TweetTexts[i]), len(TweetTexts[i])), unsafe.String(&decodedQuery[0], decodedLen)) != -1 {
				w.Body = append(w.Body, TweetHTMLs[i]...)
			}
		}
	} else {
		w.Code = HTTPStatusOK
		w.Body = append(w.Body, IndexPageFull...)
	}
	w.Body = append(w.Body, PageFinisher...)
}

func TweetPageHandler(w *HTTPResponse, r *HTTPRequest) {
	id, ok := StrToPositiveInt(r.URL.Path[len("/tweet/"):])
	if (!ok) || (id < 0) || (id > len(TweetHTMLs)-1) {
		w.Code = HTTPStatusBadRequest
		return
	}

	w.Code = HTTPStatusOK
	w.Body = append(w.Body, *TweetPage...)
	w.Body = append(w.Body, TweetHTMLs[id]...)
	w.Body = append(w.Body, PageFinisher...)
}

func PhotoHandler(w *HTTPResponse, r *HTTPRequest) {
	w.Code = HTTPStatusOK
	w.ContentType = "image/png"
	w.Body = append(w.Body, *Photo...)
}

func Router(w *HTTPResponse, r *HTTPRequest) {
	if r.URL.Path == "/" {
		IndexPageHandler(w, r)
	} else if (len(r.URL.Path) == len("/photo.jpg")) && (r.URL.Path == "/photo.jpg") {
		PhotoHandler(w, r)
	} else if (len(r.URL.Path) == len("/favicon.ico")) && (r.URL.Path == "/favicon.ico") {
		/* Do nothing :) */
	} else if (len(r.URL.Path) > len("/tweet/")) && (r.URL.Path[:len("/tweet/")] == "/tweet/") {
		TweetPageHandler(w, r)
	} else {
		w.Code = HTTPStatusNotFound
	}
}

func main() {
	IndexPage = ReadPage("pages/index.html")
	TweetPage = ReadPage("pages/tweet.html")
	Photo = ReadPage("pages/photo.jpg")
	RSSPage = ReadPage("pages/rss.xml")
	go MonitorPages()

	ReadTweets()
	go MonitorTweets()

	ConstructIndexPage()
	ConstructRSSPage()

	const port = 7070
	println("Listening on 0.0.0.0:7070...")
	if err := HTTPListenAndServe(port, Router); err != nil {
		Fatal(err.Error(), 0)
	}
}
