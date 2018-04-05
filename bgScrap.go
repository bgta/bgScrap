package main

import (
	"strconv"
	"time"

	// import standard libraries
	"fmt"
	"net/http"
	URL "net/url"
	"strings"
	// import third party libraries

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

const _defaultAuthorName string = "Raul Romero Garcia"
const _defaultAuthorEmail string = "raul@bgta.net"

// InvalidURLError is an error type when MilTorrent URL is invalid
type InvalidURLError string

func (e InvalidURLError) Error() string {
	return "Invalid URL " + strconv.Quote(string(e))
}

var feed *feeds.Feed

// Generate a Feed object
func generateFeed(title string, link string, description string) *feeds.Feed {
	return &feeds.Feed{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Description: description,
		Author:      &feeds.Author{Name: _defaultAuthorName, Email: _defaultAuthorEmail},
		Created:     time.Now(),
	}
}

func generateFeedItem(title string, link string, description string) *feeds.Item {
	return &feeds.Item{
		Title:       title,
		Link:        &feeds.Link{Href: link},
		Description: description,
		Author:      &feeds.Author{Name: _defaultAuthorName, Email: _defaultAuthorEmail},
		Created:     time.Now(),
	}
}

func fetchTorrentLinks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url := vars["q"]
	// region Solves too many open files error
	// http://craigwickesser.com/2015/01/golang-http-to-many-open-files/
	r.Close = true
	// endregion
	if url == "" {
		fmt.Println("(EE) No url specified")
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if strings.Contains(url, "%3A") {
		newURL, err := URL.QueryUnescape(url)
		if err != nil {

		}
		url = newURL
	}

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("(EE) Error: %s\n", err)
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}
	cookies := resp.Cookies()
	resp.Body.Close()

	if strings.Contains(url, "miltorrents.com") || strings.Contains(url, "mastorrents") {
		feed, err = generateMilTorrentsFeed(url, cookies)
		if err != nil {
			fmt.Printf("(EE) Error: %s\n", err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	} else if strings.Contains(url, "elitetorrent") {
		feed, err = generateEliteTorrentFeed(url, cookies)
		if err != nil {
			fmt.Printf("(EE) Error: %s\n", err)
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}
	} else {
		c := colly.NewCollector(colly.Async(false))
		c.SetCookies(url, cookies)

		c.OnHTML("title", func(e *colly.HTMLElement) {
			feed = generateFeed(e.Text, url, "")
			feed.Items = make([]*feeds.Item, 0)
		})

		c.OnHTML("a[href]", func(e *colly.HTMLElement) {
			link := e.Attr("href")
			if strings.HasSuffix(link, "torrent") {
				item := generateFeedItem(link, link, feed.Title)
				feed.Items = append(feed.Items, item)
			}
		})

		//putHeader(w, url)
		fmt.Printf("(II) Fetching %s ... ", url)
		c.Visit(url)
	}

	//putFooter(w)
	if feed != nil {
		atom, err := feed.ToRss()
		if err != nil {
			fmt.Printf("\n(EE) Error: %s\n", err)
		} else {
			fmt.Printf("%d items processed\n", len(feed.Items))
			w.Header().Set("Content-Type", "application/rss+xml")
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, atom)
		}
	} else {
		fmt.Printf("\n(EE) Error: %s\n", "No torrents found")
	}

}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/scrap", fetchTorrentLinks).Queries("q", "{q}")
	fmt.Println("Starting server on :3000")
	http.ListenAndServe(":3000", r)
}
