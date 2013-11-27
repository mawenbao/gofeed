package main

import(
    "encoding/xml"
    "log"
    "net/http"
)

func FeedEntryToRss2Item(entry *FeedEntry) (item Rss2Item) {
    item.Title = entry.Title
    item.Link = entry.Link.String()
    item.Description = string(entry.Content)
    item.PubDate = entry.Cache.LastModified.Format(http.TimeFormat)
    item.Guid = entry.Link.String()

    return
}

func GenerateRss2Feed(feed *Feed) (rss2FeedStr []byte, err error) {
    rss2Feed := &Rss2Feed { Version: FEED_VERSION }
    rss2Feed.Channel = Rss2Channel {
        Title: feed.Title,
        Link: feed.URL.String(),
        Description: "",
        PubDate: feed.LastModified.Format(http.TimeFormat),
        Generator: GOFEED_NAME + " " + GOFEED_VERSION,
    }

    rss2Feed.Channel.Items = make([]Rss2Item, len(feed.Entries))
    for itemInd, entry := range feed.Entries {
        rss2Feed.Channel.Items = append(rss2Feed.Channel.Items[:itemInd], FeedEntryToRss2Item(entry))
    }

    rss2FeedStr, err = xml.MarshalIndent(rss2Feed, "  ", "    ")
    if nil != err {
        log.Printf("[ERROR] failed to marshal rss2 feed: %s", err)
    }

    return
}

