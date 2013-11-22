package main

import (
    "strings"
    "log"
    "regexp"
)

var(
    PATTERN_ALL = [...]string { PATTERN_ANY, PATTERN_CONTENT, PATTERN_LINK, PATTERN_TITLE }
    PATTERN_ALL_REGEX = regexp.MustCompile(strings.Join(PATTERN_ALL[:], "|"))

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

func ParseIndexHtml(tar Target, htmlData []byte) (feed *Feed) {
    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        log.Printf("no match for target %s\n", tar.URL)
        return
    }

    indexRegStr := PatternToRegex(tar.IndexPattern)
    indexReg, err := regexp.Compile(indexRegStr)
    if nil != err {
        log.Printf("failed to index compile regular expression %s\n", indexRegStr)
        return
    }

    match := indexReg.FindSubmatch(htmlData)
    if nil == match {
        log.Printf("failed to match index html %s, pattern %s did not match\n", tar.URL, indexRegStr)
        return
    }

    if 3 != len(match) {
        log.Printf("failed to find regexp, got %s\n", strings.Join(indexReg.SubexpNames(), " AND "))
        return
    }

    for _, m := range match {
        println("===")
        println(string(m))
        println("===")
    }

    println("LENGTH", len(match), len(indexReg.SubexpNames()))

    for i, patName := range indexReg.SubexpNames() {
        // skip whole match 
        if 0 == i {
            continue
        }
        // no anonymous group
        if "" == patName {
            log.Printf("encountered anonymous group in pattern %s", indexRegStr)
            return
        } else if TITLE_NAME == patName {
            feed.Title = string(match[i])
        } else if LINK_NAME == patName {
            feed.Link = string(match[i])
        }
    }

    if "" == feed.Title || "" == feed.Link {
        log.Printf("empty title <%s> or empty link <%s>\n", feed.Title, feed.Link)
        return nil
    }

    return
}

func ParseContentHtml(tar Target, htmlData []byte, feed *Feed) {
    if !FilterHtmlWithoutPattern(htmlData, tar.ContentPattern) {
        log.Printf("no match for target %s\n", tar.URL)
        return
    }

    contentRegStr := PatternToRegex(tar.ContentPattern)
    contentReg, err := regexp.Compile(contentRegStr)
    if nil != err {
        log.Printf("failed to compile content regular expression %s\n", contentRegStr)
        return
    }

    match := contentReg.FindSubmatch(htmlData)
    if nil == match {
        log.Printf("failed to match content html %s, pattern %s match failed\n", tar.URL, contentRegStr)
    }

    if 2 != len(match) {
        log.Printf("failed to find regexp, got %s\n", strings.Join(contentReg.SubexpNames(), " AND "))
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
            feed.Content = match[i]
            return
        }
    }

    log.Printf("empty content for feed %s", feed.Title)
    return
}

