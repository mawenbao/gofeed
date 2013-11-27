package main

import(
    "encoding/xml"
    "log"
    "time"
    "net/http"
)

func FeedEntryToRss2Item(entry *FeedEntry) (item Rss2Item) {
    item.Title = entry.Title
    item.Link = entry.Link
    item.Description = string(entry.Content)
    if "" == entry.Cache.LastModified {
        item.PubDate = time.Now().Format(http.TimeFormat)
    } else {
        item.PubDate = entry.Cache.LastModified
    }
    item.Guid = entry.Link

    return
}

func GenerateRss2Feed(indexCache HtmlCache, entries []FeedEntry) (feedStr []byte, err error) {
    if "" == indexCache.LastModified {
        indexCache.LastModified = time.Now().Format(http.TimeFormat)
    }

    feed := Rss2Feed { Version: FEED_VERSION }
    feed.Channel = Rss2Channel {
        Title: indexCache.Title,
        Link: indexCache.URL,
        Description: "",
        PubDate: indexCache.LastModified,
        Generator: GOFEED_NAME + " " + GOFEED_VERSION,
    }

    feed.Channel.Items = make([]Rss2Item, len(entries))
    for itemInd, entry := range entries {
        feed.Channel.Items = append(feed.Channel.Items[:itemInd], FeedEntryToRss2Item(&entry))
    }

    feedStr, err = xml.MarshalIndent(&feed, "  ", "    ")
    if nil != err {
        log.Printf("failed to marshal rss2 feed: %s", err)
    }

    return
}

