package main

import (
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	GOFEED_NAME    = "gofeed"
	GOFEED_VERSION = "0.1.2"
	GOFEED_PROJECT = "https://github.com/mawenbao/gofeed"

	// used to normalize urls
	HTTP_SCHEME  = "http://"
	HTTPS_SCHEME = "https://"

	// feed related
	FEED_TYPE    = "rss"
	FEED_VERSION = "2.0"

	// used for extracting feed title/link/content
	HTML_TITLE_REG = `(?s)<\s*?html.*?<\s*?head.*?<\s*?title\s*?>(?P<title>.+)</\s*?title`
	// for cache life time
	CACHE_LIFETIME_ALL_REG = `^([1-9][0-9]*[smhd])+$`
	CACHE_LIFETIME_REG     = `([1-9][0-9]*)([smhd])`

	TITLE_NAME        = "title"
	PATTERN_TITLE     = "{" + TITLE_NAME + "}"
	PATTERN_TITLE_REG = `(?P<` + TITLE_NAME + `>(?s).+?)`

	LINK_NAME        = "link"
	PATTERN_LINK     = "{" + LINK_NAME + "}"
	PATTERN_LINK_REG = `(?P<` + LINK_NAME + `>(?s).+?)`

	CONTENT_NAME        = "description"
	PATTERN_CONTENT     = "{" + CONTENT_NAME + "}"
	PATTERN_CONTENT_REG = `(?P<` + CONTENT_NAME + `>(?s).*?)`

	PUBDATE_NAME        = "pubdate"
	PATTERN_PUBDATE     = "{" + PUBDATE_NAME + "}"
	PATTERN_PUBDATE_REG = `(?P<` + PUBDATE_NAME + `>(?s).*?)`

	FILTER_NAME        = "filter"
	PATTERN_FILTER     = "{" + FILTER_NAME + "}"
	PATTERN_FILTER_REG = `(?P<` + FILTER_NAME + `>(?s).+?)`

	PATTERN_ANY     = "{any}"
	PATTERN_ANY_REG = "(?s).*?"

	// db related consts
	DB_DRIVER           = "sqlite3"
	DB_NAME             = "cache.db"
	DB_HTML_CACHE_TABLE = "html_cache"
)

var (
	// used to set http client header User-Agent
	GOFEED_AGENT = fmt.Sprintf("Mozilla/5.0 (compatible; %s/%s; +%s)", GOFEED_NAME, GOFEED_VERSION, GOFEED_PROJECT)

	// used for filtering html
	PATTERN_ALL       = []string{PATTERN_ANY, PATTERN_CONTENT, PATTERN_LINK, PATTERN_TITLE}
	PATTERN_ALL_REGEX = regexp.MustCompile(strings.Join(PATTERN_ALL, "|"))

	// used for minifying html
	HTML_WHITESPACE_REGEX  = regexp.MustCompile(`>\s+`)
	HTML_WHITESPACE_REGEX2 = regexp.MustCompile(`\s+<`)
	HTML_WHITESPACE_REPL   = []byte(">")
	HTML_WHITESPACE_REPL2  = []byte("<")

	// used for removing junk entry content
	HTML_SCRIPT_TAG = regexp.MustCompile(`<script(?s).*?</script>`)

	// time related stuff
	GOFEED_DEFAULT_TIMEZONE, _ = time.LoadLocation("Asia/Shanghai")
)

type Config struct {
	CacheDB       string         `json:"CacheDB"`
	CacheLifetime string         `json:"CacheLifetime"` // "" means cache lives forever
	Targets       []TargetConfig `json:"Targets"`
}

type TargetConfig struct {
	Title                 string        `json:"Feed.Title"`
	Description           string        `json:"Feed.Description"`
	URLs                  []string      `json:"Feed.URL"`
	IndexPatterns         []string      `json:"Feed.IndexPattern"`
	ContentPatterns       []string      `json:"Feed.ContentPattern"`
	IndexFilterPatterns   []string      `json:"Feed.IndexFilterPattern"`
	ContentFilterPatterns []string      `json:"Feed.ContentFilterPattern"`
	PubDateFormats        []string      `json:"Feed.PubDateFormat"`
	FeedPath              string        `json:"Feed.Path"`
	ReqInterval           time.Duration `json:"Request.Interval"`
}

type FeedTarget struct {
	Title             string
	Description       string
	URLs              []*url.URL
	IndexRegs         []*regexp.Regexp
	ContentRegs       []*regexp.Regexp
	IndexFilterRegs   []*regexp.Regexp
	ContentFilterRegs []*regexp.Regexp
	FeedPath          string
	ReqInterval       time.Duration
	CacheDB           string
	CacheLifetime     time.Duration
	PubDateFormats    []string
}

type Feed struct {
	Title        string
	Description  string
	URL          *url.URL   // URL == nil means feed is invalid
	LastModified *time.Time // pubDate, cannot be nil
	Entries      []*FeedEntry
}

type FeedEntry struct {
	IndexPattern *regexp.Regexp
	Title        string
	Link         *url.URL // Link == nil means entry is invalid
	PubDate      *time.Time
	Content      []byte     // entry description
	Cache        *HtmlCache // Cache == nil means entry is invalid
}

type Rss2Feed struct {
	XMLName xml.Name    `xml:"rss"`
	Version string      `xml:"version,attr"`
	Channel Rss2Channel `xml:"channel"`
}

type Rss2Channel struct {
	Title       string     `xml:"title"`
	Link        string     `xml:"link"`
	Description string     `xml:"description"`
	PubDate     string     `xml:"pubDate"`
	Generator   string     `xml:"generator"`
	Items       []Rss2Item `xml:"item"`
}

type Rss2Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description",chardata`
	PubDate     string `xml:"pubDate"`
	Guid        string `xml:"guid"`
}

const (
	CACHE_NOT_MODIFIED = iota
	CACHE_NEW
	CACHE_MODIFIED
)

type HtmlCache struct {
	Status int // default is CACHE_NOT_MODIFIED

	URL          *url.URL
	Date         *time.Time // date of the html request, should never be nil
	CacheControl string
	LastModified *time.Time
	Etag         string
	Expires      *time.Time
	Html         []byte
}

// query returns emtpy record set
type DBNoRecordError struct {
}

func (nre DBNoRecordError) Error() string {
	return "db query returned empty record set"
}
