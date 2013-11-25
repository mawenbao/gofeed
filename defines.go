package main

import(
    "strings"
    "time"
    "regexp"
)

const(
    SCHEME_SUFFIX = "://"
    HTTP_SCHEME = "http"

    TITLE_NAME = "title"
    PATTERN_TITLE = "{" + TITLE_NAME + "}"
    PATTERN_TITLE_REG = `(?P<` + TITLE_NAME + `>(?s).+?)`

    LINK_NAME = "link"
    PATTERN_LINK = "{" + LINK_NAME + "}"
    PATTERN_LINK_REG = `(?P<` + LINK_NAME + `>(?s).+?)`

    CONTENT_NAME = "content"
    PATTERN_CONTENT = "{" + CONTENT_NAME + "}"
    PATTERN_CONTENT_REG = "(?P<" + CONTENT_NAME + ">(?s).*?)"

    PATTERN_ANY = "{any}"
    PATTERN_ANY_REG = "(?s).*?"
)

var(
    // used for filtering html
    PATTERN_ALL = []string { PATTERN_ANY, PATTERN_CONTENT, PATTERN_LINK, PATTERN_TITLE }
    PATTERN_ALL_REGEX = regexp.MustCompile(strings.Join(PATTERN_ALL, "|"))

    // used for minifying html
    HTML_WHITESPACE_REGEX = regexp.MustCompile(`>\s+`)
    HTML_WHITESPACE_REGEX2 = regexp.MustCompile(`\s+<`)
    HTML_WHITESPACE_REPL = []byte(">")
    HTML_WHITESPACE_REPL2 = []byte("<")
)

type Target struct {
    URL string `json:"Feed.URL"`
    IndexPattern string `json:"Feed.IndexPattern"`
    ContentPattern string `json:"Feed.ContentPattern"`
    Path string `json:"Feed.Path"`
    ReqInterval time.Duration `json:"Request.Interval"`
}

type TargetSlice struct {
    Targets []Target
}

type FeedEntry struct {
    Title string
    Link string
    Content []byte
}

