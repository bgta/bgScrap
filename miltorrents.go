package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gocolly/colly"
	"github.com/gorilla/feeds"
)

func generateMilTorrentsFeed(urlBase string, cookies []*http.Cookie) (*feeds.Feed, error) {
	re := regexp.MustCompile("^(.*)-([0-9]+)x([0-9]+)$")
	matches := re.FindStringSubmatch(urlBase)
	var feed *feeds.Feed

	if len(matches) == 4 {
		for i := 1; i < 99; i++ {
			counter := 0
			url := matches[1] + "-" + matches[2] + fmt.Sprintf("x%02d", i)
			c := colly.NewCollector(colly.Async(false))
			c.SetCookies(url, cookies)
			c.OnHTML("title", func(e *colly.HTMLElement) {
				if feed == nil {
					feed = generateFeed(e.Text, url, "")
					feed.Items = make([]*feeds.Item, 0)
				}
			})

			c.OnHTML("a[href]", func(e *colly.HTMLElement) {
				link := e.Attr("href")
				if strings.Contains(link, ".torrent") && !strings.Contains(link, "720") {
					rei := regexp.MustCompile("showDownload.'(.*)','(.*)','(.*)','(.*).torrent'.")
					matchesi := rei.FindStringSubmatch(link)
					if len(matchesi) == 5 {
						counter++
						link = matchesi[4] + ".torrent"
						item := generateFeedItem(matchesi[2], link, matchesi[2]+" - "+matchesi[1])
						feed.Items = append(feed.Items, item)
					}

				}
			})

			fmt.Printf("(II) [Miltorrents] Fetching %s ... \n", url)
			c.Visit(url)
			if counter == 0 {
				break
			} else {
				fmt.Printf("(II) [Miltorrents] %d torrents fetched from %s ... \n", counter, url)
			}
		}

		return feed, nil

	}

	return nil, InvalidURLError(urlBase)
}
