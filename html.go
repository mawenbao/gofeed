package main

import (
    "strings"
    "log"
    "regexp"
    "time"
)

func ExtractHtmlTitle(htmlData []byte) string {
    titleReg := regexp.MustCompile(HTML_TITLE_REG)
    matches := titleReg.FindSubmatch(htmlData)
    if len(matches) != 2 {
        return ""
    }
    return string(matches[1])
}

func MinifyHtml(htmlData []byte) []byte {
    htmlData = HTML_WHITESPACE_REGEX.ReplaceAll(htmlData, HTML_WHITESPACE_REPL)
    htmlData = HTML_WHITESPACE_REGEX2.ReplaceAll(htmlData, HTML_WHITESPACE_REPL2)
    return htmlData
}

func FilterHtmlWithoutPattern(htmlData []byte, pattern string) bool {
    html := string(htmlData)
    for _, str := range PATTERN_ALL_REGEX.Split(pattern, -1) {
        if "" == str {
            continue
        }
        if !strings.Contains(html, str) {
            log.Printf("[ERROR] target html does not contain %s", str)
            return false
        }
    }

    return true
}

func ParseIndexHtml(feedTar *FeedTarget) (feed *Feed, ok bool) {
    feed = new(Feed)

    for urlInd, tarURL := range feedTar.URLs {
        // get cache first
        indexCache, err := FetchHtml(tarURL, feedTar.CacheDB)
        if nil != err {
            log.Printf("[ERROR] failed to download index web page %s", tarURL.String())
            // just ignore the sucker
            continue
        }

        // minify html
        htmlData := MinifyHtml(indexCache.Html)

        // extract feed entry title and link
        indexReg := FindIndexReg(feedTar, tarURL)
        if nil == indexReg {
            log.Printf("[ERROR] cannot find index regex for %s", tarURL.String)
            continue
        }
        matches := indexReg.FindAllSubmatch(htmlData, -1)
        if nil == matches {
            log.Printf("[ERROR] failed to match index html %s, pattern %s did not match", tarURL.String(), indexReg.String())
            // ignore the sucker
            continue
        }
        entries := make([]*FeedEntry, len(matches))
        for matchInd, match := range matches {
            entries[matchInd] = new(FeedEntry)
            entry := entries[matchInd] // pointer of FeedEntry
            for patInd, patName := range indexReg.SubexpNames() {
                switch patName {
                case TITLE_NAME:
                    entry.Title = string(match[patInd])
                case LINK_NAME:
                    // normalize entry link which may be relative
                    entry.Link, err = tarURL.Parse(string(match[patInd]))
                    if nil != err {
                        log.Printf("[ERROR] error parsing entry link %s: %s", entry.Link, err)
                    }
                }
            }
        }

        // set feed
        feed.Entries = append(feed.Entries, entries...)
        if 0 == urlInd {
            // use first index page's title and url
            feed.Title = indexCache.Title
            feed.URL = tarURL
            feed.LastModified = indexCache.LastModified
        } else {
            // use later lastmod time
            if feed.LastModified.Before(indexCache.LastModified) {
                feed.LastModified = indexCache.LastModified
            }
        }
    }
    return feed, true
}

func ParseContentHtml(feedTar *FeedTarget, feed *Feed) (ok bool) {
    for _, entry := range feed.Entries {
        // wait some time
        if *gVerbose {
            log.Printf("waiting for %d seconds before sending request to %s", feedTar.ReqInterval, entry.Link.String())
        }
        time.Sleep(feedTar.ReqInterval * time.Second)

        cache, err := FetchHtml(entry.Link, feedTar.CacheDB)
        if nil != err {
            log.Printf("[ERROR] failed to download web page %s", entry.Link.String())
            continue
        }
        entry.Cache = &cache

        htmlData := MinifyHtml(cache.Html)

        // extract feed entry content(description)
        contentReg := FindContentReg(feedTar, feed.URL)
        if nil == contentReg {
            log.Printf("[ERROR] failed to find content regex for %s", feed.URL.String())
            continue
        }
        match := contentReg.FindSubmatch(htmlData)
        if nil == match {
            log.Printf("[ERROR] failed to match content html %s, pattern %s match failed", entry.Link.String(), contentReg.String())
            // ignore this sucker
            continue
        }
        for i, patName := range contentReg.SubexpNames() {
            if CONTENT_NAME == patName {
                entry.Content = match[i]
            }
        }

        if 0 == len(entry.Content) {
            // just print a warning message if content is empty
            log.Printf("[WARN] empty content for feed entry %s", entry.Title)
        }
    }

    return true
}

