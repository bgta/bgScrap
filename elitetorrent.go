package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
)

func generateEliteTorrentFeed(url string, cookies []*http.Cookie) (*feeds.Feed, error) {
	// https://www.elitetorrent.biz/series/the-flash-temporada-4-capitulo-16/
	var feed *feeds.Feed
	url = strings.Replace(url, " ", "+", -1)
	c := colly.NewCollector(colly.Async(false))
	c.SetCookies(url, cookies)
	c.OnHTML("title", func(e *colly.HTMLElement) {
		if feed == nil {
			feed = generateFeed(e.Text, url, "")
			feed.Items = make([]*feeds.Item, 0)
		}
	})

	c.OnHTML("a.nombre[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		title := e.Attr("title")

		fmt.Printf("(II) [EliteTorrent] Fetching %s -> %s ... \n", link, title)
		if strings.Contains(link, "720p") {
			return
		}
		cb := colly.NewCollector(colly.Async(false))
		cb.SetCookies(url, cookies)
		cb.OnHTML("title", func(e *colly.HTMLElement) {
			title = e.Text
		})

		cb.OnHTML("a.enlace_torrent[href]", func(e *colly.HTMLElement) {
			torrentLink := e.Attr("href")
			if strings.HasSuffix(torrentLink, "torrent") {
				finalLink := "https://www.elitetorrent.biz" + torrentLink
				item := generateFeedItem(finalLink, finalLink, title)
				feed.Items = append(feed.Items, item)
			}
		})

		cb.Visit(link)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	fmt.Printf("(II) [EliteTorrent] Fetching %s ... \n", url)
	c.Visit(url)

	c.Wait()

	return feed, nil
}
