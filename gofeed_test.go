package main

import (
	"bytes"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	// print detailed debug infomation
	*gDebug = true
	//*gVerbose = true
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func TestExtractCacheLifetime(t *testing.T) {
	sampleCacheTimeStr := []string{
		"2d",
		"100h",
		"2D5h3M4s",
		"",
		"2d5d3s",
		"2d3M4S10s",
		"-1d",
		"3d1sabcde",
		"05m",
		"abc",
	}
	expectCacheTime := []time.Duration{
		2 * 24 * time.Hour,
		100 * time.Hour,
		(2*24+5)*time.Hour + 3*time.Minute + 4*time.Second,
		time.Duration(-1),
		time.Duration(-2),
		time.Duration(-2),
		time.Duration(-2),
		time.Duration(-2),
		time.Duration(-2),
		time.Duration(-2),
	}

	for i, _ := range sampleCacheTimeStr {
		actualCacheLife := ExtractCacheLifetime(sampleCacheTimeStr[i])
		if expectCacheTime[i] != actualCacheLife {
			t.Fatalf("extract cache lifetime failed, expect %d from %s, got %d", expectCacheTime[i], sampleCacheTimeStr[i], actualCacheLife)
		}
	}
}

func TestParseJsonConfig(t *testing.T) {
	config_file := "example_config2.json"
	feedTargets := ParseJsonConfig(config_file)
	feedTar := feedTargets[0]

	feedURL := "http://blog.atime.me"
	if feedURL != feedTar.URLs[0].String() {
		t.Fatalf("%s: failed to parse url, expected %s, got %s", config_file, feedURL, feedTar.URLs[0].String())
	}

	feedIndexPattern := `<div class="niu2-index-article-title"><span><a href="(?P<link>(?s).+?)">(?P<title>(?s).+?)</a></span></div>`
	if feedIndexPattern != feedTar.IndexRegs[0].String() {
		t.Fatalf("%s: failed to parse index pattern, expected %s, got %s", config_file, feedIndexPattern, feedTar.IndexRegs[0].String())
	}

	feedContentPattern := `<div class="clearfix visible-xs niu2-clearfix"></div>(?P<description>(?s).*?)<div id="content-comments">`
	if feedContentPattern != feedTar.ContentRegs[0].String() {
		t.Fatalf("%s: failed to parse content pattern, expected %s, got %s", config_file, feedContentPattern, feedTar.ContentRegs[0].String())
	}

	feedPath, _ := filepath.Abs(`blog.atime.me.xml`)
	if feedPath != feedTar.FeedPath {
		t.Fatalf("%s: failed to parse path, expected %s, got %s", config_file, feedPath, feedTar.FeedPath)
	}
}

/*
func TestRequestHtml(t *testing.T) {
    url := "http://blog.atime.me/agreement.html"
    cache := HtmlCache { URL: url }
    err := requestHtml(&cache)
    if nil != err {
        t.Fatalf("failed to crawl %s: %s", url, err)
    }

    testFile := "test_data/test_crawl.html"
    expectData, err := ioutil.ReadFile(testFile)

    if nil != err {
        t.Fatalf("failed to read %s: %s", testFile, err)
    }

    if 0 != bytes.Compare(expectData, cache.Html) {
        t.Fatalf("html data crawled from %s not equal to %s", url, testFile)
    }

}
*/

func TestFetchHtml(t *testing.T) {
	feedTargets := ParseJsonConfig("example_config2.json")
	cacheDB := feedTargets[0].CacheDB

	os.Remove(cacheDB)
	err := CreateDBScheme(cacheDB)
	if nil != err {
		t.Fatalf("failed to create db at %s: %s", cacheDB, err)
	}
	defer os.Remove(cacheDB)

	// new cache
	url, _ := url.Parse("http://blog.atime.me/agreement.html")
	cache, err := FetchHtml(url, cacheDB, time.Duration(-1))
	if nil != err {
		t.Fatalf("failed to fetch html %s", err)
	}

	if url != cache.URL {
		t.Fatalf("wrong html cache, url not match")
	}

	cache2, err := GetHtmlCacheByURL(cacheDB, url.String())
	if nil != err {
		t.Fatalf("html cache not saved for url %s", url)
	}

	if cache.URL.String() != cache2.URL.String() ||
		!cache.LastModified.Equal(*cache2.LastModified) ||
		0 != bytes.Compare(cache.Html, cache2.Html) {

		t.Logf("url %t, %s vs %s", cache.URL.String() == cache2.URL.String(), cache.URL.String(), cache2.URL.String())
		t.Logf("lastmod %t, %s vs %s", cache.LastModified.Equal(*cache2.LastModified), cache.LastModified.String(), cache2.LastModified.String())
		t.Logf("html length %t, %d vs %d", len(cache.Html) == len(cache2.Html), len(cache.Html), len(cache2.Html))
		t.Fatalf("html cache not match")
	}

	// use old cache
	cache4, err := FetchHtml(url, cacheDB, time.Duration(-1))
	if nil != err || CACHE_NOT_MODIFIED != cache4.Status {
		t.Fatalf("failed to reuse html cache for %s: %s", url, err)
	}
}

func TestCheckPatterns(t *testing.T) {
	invalidTargets := [...]TargetConfig{
		TargetConfig{IndexPatterns: []string{""}, ContentPatterns: []string{""}},
		TargetConfig{IndexPatterns: []string{"abc"}, ContentPatterns: []string{""}},
		TargetConfig{IndexPatterns: []string{"abc"}, ContentPatterns: []string{"cde"}},
		TargetConfig{IndexPatterns: []string{"abc{link}"}, ContentPatterns: []string{"cde"}},
		TargetConfig{IndexPatterns: []string{"{title} {link}abc{title}"}, ContentPatterns: []string{"{description}"}},
		TargetConfig{IndexPatterns: []string{"{*}abc{title}"}, ContentPatterns: []string{"{title}"}},
		TargetConfig{IndexPatterns: []string{"{link}abc{title}"}, ContentPatterns: []string{"{title}{description}"}},
		TargetConfig{IndexPatterns: []string{"{link}abc{title}"}, ContentPatterns: []string{"{link}{*}{description}"}},
	}

	validTargets := [...]TargetConfig{
		TargetConfig{IndexPatterns: []string{"{link}abc{title}"}, ContentPatterns: []string{"{*}{description}"}},
		TargetConfig{IndexPatterns: []string{"{link}abc{*}cde{title}"}, ContentPatterns: []string{"{description}"}},
	}

	for _, tar := range invalidTargets {
		if CheckPatterns(&tar) {
			t.Fatal("check patterns failed")
		}
	}

	for _, tar := range validTargets {
		if !CheckPatterns(&tar) {
			t.Fatal("check pattern failed")
		}
	}
}

func TestExtractHtmlTitle(t *testing.T) {
	blogURL, _ := url.Parse("http://blog.atime.me")
	cache := HtmlCache{URL: blogURL}
	resp, err := SendHttpRequest(&cache)
	if nil != err {
		t.Fatalf("failed to send http request to %s", cache.URL.String())
	}
	if err = ParseHttpResponse(resp, &cache); nil != err {
		t.Fatalf("failed to parse http response for %s", cache.URL.String())
	}

	expectedTitle := "MWB日常笔记"
	realTitle := ExtractHtmlTitle(cache.Html)
	if expectedTitle != realTitle {
		t.Fatalf("title mismatch, expected %s, got %s", expectedTitle, realTitle)
	}
}

func TestMinifyHtml(t *testing.T) {
	rawHtml := `<html>
    <head>  </head>  
  <body> Hello  world
</body>
</html>`
	expectedHtml := "<html><head></head><body>Hello  world</body></html>"
	if expectedHtml != string(MinifyHtml([]byte(rawHtml))) {
		t.Fatal("failed to minify html")
	}
}

/*
func TestFilterHtmlWithoutPattern(t *testing.T) {
    feedTargets := ParseJsonConfig("example_config2.json")
    cache := HtmlCache { URL: tar.URL }
    err = RequestHtml(&cache)
    if nil != err {
        t.Fatal("failed to download web page")
    }
    htmlData := MinifyHtml(cache.Html)

    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        t.Fatalf("filter without index pattern failed for target %s", tar.URL)
    }
}
*/

func TestDB(t *testing.T) {
	feedTargets := ParseJsonConfig("example_config2.json")
	cacheDB := feedTargets[0].CacheDB

	os.Remove(cacheDB)

	err := CreateDBScheme(cacheDB)
	if nil != err {
		t.Fatalf("failed to create db %s: %s", cacheDB, err)
	}
	defer os.Remove(cacheDB)

	url1, _ := url.Parse("http://blog.atime.me")
	url2, _ := url.Parse("http://atime.me")

	dateNow := time.Now()
	cache := []*HtmlCache{
		&HtmlCache{URL: url1, Date: &dateNow, LastModified: &dateNow, Html: []byte("hello world")},
		&HtmlCache{URL: url2, Date: &dateNow, LastModified: &dateNow, Html: []byte("hello world")},
	}

	err = PutHtmlCache(cacheDB, cache)
	if nil != err {
		t.Fatalf("failed to insert records to db %s: %s", cacheDB, err)
	}

	cache2, err := GetHtmlCacheByURL(cacheDB, "no.cache")
	if nil == err {
		t.Fatalf("should not get html cache from an non-exist url")
	}

	cache2, err = GetHtmlCacheByURL(cacheDB, cache[0].URL.String())
	if nil != err {
		t.Fatalf("failed to get html cache for url %s: %s", cache[0].URL.String(), err)
	}
	if cache2.URL.String() != cache[0].URL.String() ||
		cache2.LastModified.Format(http.TimeFormat) != cache[0].LastModified.Format(http.TimeFormat) ||
		0 != bytes.Compare(cache2.Html, cache[0].Html) {

		t.Fatalf("got wrong html cache")
	}

	// update db
	cache2.CacheControl = "ok, I know this is not true"
	err = UpdateHtmlCache(cacheDB, []*HtmlCache{cache2})
	if nil != err {
		t.Fatalf("failed to update db: %s", err)
	}

	cache3, err := GetHtmlCacheByURL(cacheDB, cache2.URL.String())
	if cache2.CacheControl != cache3.CacheControl {
		t.Fatalf("updated CacheControl does match, %s vs %s", cache2.CacheControl, cache3.CacheControl)
	}

	// test remove
	err = DelHtmlCacheByURL(cacheDB, cache2.URL.String())
	if nil != err {
		t.Fatal("failed to remove record from db")
	}
	_, err = GetHtmlCacheByURL(cacheDB, cache2.URL.String())
	if nil == err {
		t.Fatal("failed to remove record from db, record still exists")
	}
}

func TestGenerateRss2Feed(t *testing.T) {
	feedTargets := ParseJsonConfig("example_config2.json")
	cacheDB := feedTargets[0].CacheDB

	os.Remove(cacheDB)
	err := CreateDBScheme(cacheDB)
	if nil != err {
		t.Fatalf("failed to create db at %s: %s", cacheDB, err)
	}
	defer os.Remove(cacheDB)

	feedTar := feedTargets[0]
	feed, ok := ParseIndexHtml(feedTar)
	if !ok {
		t.Fatalf("failed to parse index html for feed target %s", feedTar.FeedPath)
	}

	if !ParseContentHtml(feedTar, feed) {
		t.Fatalf("failed to parse content html for feed target %s", feedTar.FeedPath)
	}

	_, err = GenerateRss2Feed(feed)
	if nil != err {
		t.Fatalf("failed to generate rss")
	}
}

func TestParseIndexAndContentHtml(t *testing.T) {
	feedTargets := ParseJsonConfig("example_config2.json")
	cacheDB := feedTargets[0].CacheDB

	os.Remove(cacheDB)
	err := CreateDBScheme(cacheDB)
	if nil != err {
		t.Fatalf("failed to create db at %s: %s", cacheDB, err)
	}
	defer os.Remove(cacheDB)

	for _, tar := range feedTargets {
		feed, ok := ParseIndexHtml(tar)
		if !ok {
			t.Fatalf("failed to parse index html for feed target %s", tar.FeedPath)
		}

		if !ParseContentHtml(tar, feed) {
			t.Fatalf("failed to parse content html for feed target %s", tar.FeedPath)
		}

		for _, entry := range feed.Entries {
			println("title", entry.Title)
			println("link", entry.Link)
			println("content length", len(entry.Content))
			var entryDesc []rune = []rune(string(entry.Content))
			if len(entryDesc) > 100 {
				println("content summary", string(entryDesc)[:100])
			} else {
				println("content", string(entryDesc))
			}
		}
	}
}
