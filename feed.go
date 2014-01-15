package main

import (
	"encoding/xml"
	"errors"
	"log"
	"time"
)

func FeedEntryToRss2Item(entry *FeedEntry) (item Rss2Item) {
	if nil == entry || nil == entry.Link || nil == entry.Cache {
		log.Println("[ERROR] got invalid entry: entry is nil or entry.Link is nil or entry.Cache is nil")
		return
	}

	item.Title = entry.Title
	item.Link = entry.Link.String()
	item.Description = string(entry.Content)
	if nil != entry.PubDate {
		item.PubDate = entry.PubDate.Format(time.RFC1123Z)
	} else if nil != entry.Cache.LastModified {
		item.PubDate = entry.Cache.LastModified.Format(time.RFC1123Z)
	} else if nil != entry.Cache.Date {
		item.PubDate = entry.Cache.Date.Format(time.RFC1123Z)
	} else {
		log.Printf("[ERROR] entry's cache date is nil %s", entry.Link.String())
		item.PubDate = time.Now().Format(time.RFC1123Z)
	}
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
		PubDate:     feed.LastModified.Format(time.RFC1123Z),
		Generator:   GOFEED_NAME + " " + GOFEED_VERSION,
	}

	rss2Feed.Channel.Items = make([]Rss2Item, len(feed.Entries))
	itemInd := 0
	for entryInd, entry := range feed.Entries {
		if nil == entry {
			log.Printf("[ERROR] got nil entry at index %d", entryInd)
		} else if nil == entry.Link || nil == entry.Cache {
			log.Println("[WARN] Ignore invalid feed entry: link or cache is nil")
		} else if 0 == len(entry.Content) {
			log.Printf("[WARN] Ignore empty feed entry %s", entry.Link.String())
		} else {
			rss2Feed.Channel.Items = append(rss2Feed.Channel.Items[:itemInd], FeedEntryToRss2Item(entry))
			itemInd += 1
		}
	}

	rss2FeedStr, err = xml.MarshalIndent(rss2Feed, "  ", "    ")
	if nil != err {
		log.Printf("[ERROR] failed to marshal rss2 feed: %s", err)
	}

	return
}
