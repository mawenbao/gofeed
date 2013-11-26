package main

import(
    "strings"
    "time"
    "regexp"
    "encoding/xml"
)

const(
    GOFEED_NAME = "gofeed"
    GOFEED_VERSION = "0.1"

    // used to normalize urls
    HTTP_SCHEME = "http://"
    HTTPS_SCHEME = "https://"

    // feed related
    FEED_TYPE = "rss"
    FEED_VERSION = "2.0"

    // used for extracting feed title/link/content
    HTML_TITLE_REG = `(?s)<\s*?html.*?<\s*?head.*?<\s*?title\s*?>(?P<title>.+)</\s*?title`

    TITLE_NAME = "title"
    PATTERN_TITLE = "{" + TITLE_NAME + "}"
    PATTERN_TITLE_REG = `(?P<` + TITLE_NAME + `>(?s).+?)`

    LINK_NAME = "link"
    PATTERN_LINK = "{" + LINK_NAME + "}"
    PATTERN_LINK_REG = `(?P<` + LINK_NAME + `>(?s).+?)`

    CONTENT_NAME = "description"
    PATTERN_CONTENT = "{" + CONTENT_NAME + "}"
    PATTERN_CONTENT_REG = "(?P<" + CONTENT_NAME + ">(?s).*?)"

    PATTERN_ANY = "{any}"
    PATTERN_ANY_REG = "(?s).*?"

    // db related consts
    DB_DRIVER = "sqlite3"
    DB_NAME = "cache.db"
    DB_HTML_CACHE_TABLE = "html_cache"
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
    FeedPath string `json:"Feed.Path"`
    ReqInterval time.Duration `json:"Request.Interval"`
}

type TargetGroup struct {
    FeedPath string
    IndexCache HtmlCache
    Entries []FeedEntry
    Targets []Target
}

type Config struct {
    CacheDB string `json:"CacheDB"`
    Targets []Target `json:"Targets"`
}

type FeedEntry struct {
    Title string
    Link string
    Content []byte
    Cache *HtmlCache
}

type Rss2Feed struct {
    XMLName xml.Name `xml:"rss"`
    Version string `xml:"version,attr"`
    Channel Rss2Channel `xml:"channel"`
}

type Rss2Channel struct {
    Title string `xml:"title"`
    Link string `xml:"link"`
    Description string `xml:"description"`
    PubDate string `xml:"pubDate"`
    Generator string `xml:"generator"`
    Items []Rss2Item `xml:"item"`
}

type Rss2Item struct {
    Title string `xml:"title"`
    Link string `xml:"link"`
    Description string `xml:"description",chardata`
    PubDate string `xml:"pubDate"`
    Guid string `xml:"guid"`
}

const(
    CACHE_NOT_MODIFIED = iota
    CACHE_NEW
    CACHE_MODIFIED
)

type HtmlCache struct {
    Status int // default is CACHE_NOT_MODIFIED

    URL string // must be normalized by url.Parse
    Title string
    CacheControl string
    LastModified string // http.TimeFormat
    Etag string
    Expires string // http.TimeFormat
    Html []byte
}

// query returns emtpy record set
type DBNoRecordError struct {
}

func (nre DBNoRecordError) Error() string {
    return "db query returned empty record set"
}

