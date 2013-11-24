package main

import (
    "strings"
    "log"
    "regexp"
)

var(
    PATTERN_ALL = []string { PATTERN_ANY, PATTERN_CONTENT, PATTERN_LINK, PATTERN_TITLE }
    PATTERN_ALL_REGEX = regexp.MustCompile(strings.Join(PATTERN_ALL, "|"))

    HTML_WHITESPACE_REGEX = regexp.MustCompile(`>\s+`)
    HTML_WHITESPACE_REGEX2 = regexp.MustCompile(`\s+<`)
)

func MinifyHtml(htmlData []byte) []byte {
    htmlData = HTML_WHITESPACE_REGEX.ReplaceAll(htmlData, []byte(">"))
    htmlData = HTML_WHITESPACE_REGEX2.ReplaceAll(htmlData, []byte("<"))
    return htmlData
}

func FilterHtmlWithoutPattern(htmlData []byte, pattern string) bool {
    html := string(MinifyHtml(htmlData))
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

func ParseIndexHtml(tar Target) (entries []FeedEntry, ok bool) {
    htmlData, err := Crawl(tar.URL)
    if nil != err {
        log.Printf("failed to download index web page %s", tar.URL)
        return
    }

    htmlData = MinifyHtml(htmlData)

    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        log.Printf("no match for target %s", tar.URL)
        return
    }

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
                log.Printf("encountered anonymous group in pattern %s", indexRegStr)
                return
            } else if TITLE_NAME == patName {
                entry.Title = string(match[patInd])
            } else if LINK_NAME == patName {
                entry.Link = string(match[patInd])
            }
        }

        if "" == entry.Title || "" == entry.Link {
            log.Printf("empty title <%s> or empty link <%s>", entry.Title, entry.Link)
            return
        }
    }

    ok = true
    return
}

func ParseContentHtml(tar Target, entries []FeedEntry) (ok bool) {
    contentRegStr := PatternToRegex(tar.ContentPattern)

    for entryInd, _ := range entries {
        entry := &entries[entryInd]

        if "" == entry.Link {
            log.Printf("url of feed entry %s is empty", entry.Title)
            // just skip emtpy feed entry
            continue
        }

        htmlData, err := Crawl(entry.Link)
        if nil != err {
            log.Printf("failed to download web page %s", entry.Link)
            return
        }

        htmlData = MinifyHtml(htmlData)
        if !FilterHtmlWithoutPattern(htmlData, tar.ContentPattern) {
            log.Printf("no match for target %s", tar.URL)
            return
        }

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
            if CONTENT_NAME == patName {
                entry.Content = match[i]
            }
        }

        if 0 == len(entry.Content) {
            // just print a warning message if content is empty
            log.Printf("empty content for feed entry %s", entry.Title)
        }
    }

    ok = true
    return
}

