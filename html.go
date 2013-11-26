package main

import (
    "strings"
    "log"
    "regexp"
    "time"
    "net/url"
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
            log.Printf("target html does not contain %s", str)
            return false
        }
    }

    return true
}

func ParseIndexHtml(conf Config, tar Target) (indexCache HtmlCache, entries []FeedEntry, ok bool) {
    indexCache, err := FetchHtml(tar.URL, conf.CacheDB)
    if nil != err {
        log.Printf("failed to download index web page %s", tar.URL)
        return
    }

    htmlData := MinifyHtml(indexCache.Html)

    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        log.Printf("no match for target %s", tar.URL)
        return
    }

    // extract feed entry title and link
    indexRegStr := PatternToRegex(tar.IndexPattern)
    indexReg, err := regexp.Compile(indexRegStr)
    if nil != err {
        log.Printf("failed to index compile regular expression %s", indexRegStr)
        return
    }

    matches := indexReg.FindAllSubmatch(htmlData, -1)
    if nil == matches {
        log.Printf("failed to match index html %s, pattern %s did not match", tar.URL, indexRegStr)
        return
    }

    // should have checked target url when parsing json config
    baseURL, err := url.Parse(tar.URL)
    if nil != err {
        log.Printf("error parsing index url %s: %s", tar.URL, err)
        return
    }

    entries = make([]FeedEntry, len(matches))
    for matchInd, match := range matches {
        entry := &entries[matchInd]
        for patInd, patName := range indexReg.SubexpNames() {
            // skip whole match 
            if 0 == patInd {
                continue
            }
            // no anonymous group
            if "" == patName {
                log.Printf("encountered anonymous group with pattern %s", indexRegStr)
                return
            } else if TITLE_NAME == patName {
                entry.Title = string(match[patInd])
            } else if LINK_NAME == patName {
                entry.Link = string(match[patInd])
                // normalize entry link which may be relative
                var entryURL *url.URL
                entryURL, err = baseURL.Parse(entry.Link)
                if nil != err {
                    log.Printf("error parsing entry link %s: %s", entry.Link, err)
                } else {
                    entry.Link = entryURL.String()
                }
            }
        }
    }

    ok = true
    return
}

func ParseContentHtml(conf Config, tar Target, entry *FeedEntry) (ok bool) {
    // wait some time
    if *gVerbose {
        log.Printf("waiting for %d seconds before sending request to %s", tar.ReqInterval, entry.Link)
    }
    time.Sleep(tar.ReqInterval * time.Second)

    if "" == entry.Link {
        log.Printf("url of feed entry %s is empty", entry.Title)
        // just skip emtpy feed entry
        return true
    }

    cache, err := FetchHtml(entry.Link, conf.CacheDB)
    if nil != err {
        log.Printf("failed to download web page %s", entry.Link)
        return
    }
    entry.Cache = &cache

    htmlData := MinifyHtml(cache.Html)
    if !FilterHtmlWithoutPattern(htmlData, tar.ContentPattern) {
        log.Printf("no match for target %s", tar.URL)
        return
    }

    // extract feed entry content(description)
    contentRegStr := PatternToRegex(tar.ContentPattern)
    contentReg, err := regexp.Compile(contentRegStr)
    if nil != err {
        log.Printf("failed to compile content regular expression %s", contentRegStr)
        return
    }

    match := contentReg.FindSubmatch(htmlData)
    if nil == match {
        log.Printf("failed to match content html %s, pattern %s match failed", entry.Link, contentRegStr)
        return
    }

    for i, patName := range contentReg.SubexpNames() {
        // skip whole match 
        if 0 == i {
            continue
        }
        // no anonymous group
        if "" == patName {
            log.Printf("encountered anonymous group in pattern %s", contentRegStr)
            return
        } else if CONTENT_NAME == patName {
            entry.Content = match[i]
        }
    }

    if 0 == len(entry.Content) {
        // just print a warning message if content is empty
        log.Printf("empty content for feed entry %s", entry.Title)
    }

    ok = true
    return
}

