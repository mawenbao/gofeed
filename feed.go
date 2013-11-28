package main

import (
	"encoding/xml"
	"errors"
	"log"
	"net/http"
)

func FeedEntryToRss2Item(entry *FeedEntry) (item Rss2Item) {
	if nil == entry || nil == entry.Link || nil == entry.Cache {
		log.Println("[ERROR] got invalid entry: entry is nil or entry.Link is nil or entry.Cache is nil")
		return
	}

	item.Title = entry.Title
	item.Link = entry.Link.String()
	item.Description = string(entry.Content)
	item.PubDate = entry.Cache.LastModified.Format(http.TimeFormat)
	item.Guid = entry.Link.String()

	return
}

func GenerateRss2Feed(feed *Feed) (rss2FeedStr []byte, err error) {
	if nil == feed || nil == feed.URL {
		log.Println("[ERROR] Got empty feed, wll ignore it")
		err = errors.New("Empty feed")
		return
	}

	rss2Feed := &Rss2Feed{Version: FEED_VERSION}
	rss2Feed.Channel = Rss2Channel{
		Title:       feed.Title,
		Link:        feed.URL.String(),
		Description: "",
		PubDate:     feed.LastModified.Format(http.TimeFormat),
		Generator:   GOFEED_NAME + " " + GOFEED_VERSION,
	}

	rss2Feed.Channel.Items = make([]Rss2Item, len(feed.Entries))
	for itemInd, entry := range feed.Entries {
		if nil == entry {
			log.Println("[ERROR] got nil entry at index %d", itemInd)
		} else if nil == entry.Link || nil == entry.Cache {
			log.Printf("[WARN] Ignore invalid feed entry %s: link or cache is nil")
		} else {
			rss2Feed.Channel.Items = append(rss2Feed.Channel.Items[:itemInd], FeedEntryToRss2Item(entry))
		}
	}

	rss2FeedStr, err = xml.MarshalIndent(rss2Feed, "  ", "    ")
	if nil != err {
		log.Printf("[ERROR] failed to marshal rss2 feed: %s", err)
	}

	return
}
