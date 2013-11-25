package main

import(
    "testing"
    "os"
    "bytes"
)

func TestParseJsonConfig(t *testing.T) {
    config_file := "example_config.json"
    conf, err := ParseJsonConfig(config_file)

    if nil != err {
        t.Fatalf("Failed to parse json config file %s: %s", config_file, err)
    }

    feedTar := conf.Targets[0]

    feedURL := `blog.atime.me`
    if feedURL != feedTar.URL {
        t.Fatalf("%s: failed to parse url, expected %s, got %s", config_file, feedURL, feedTar.URL)
    }

    feedIndexPattern := `<div class="niu2-index-article-title"><span><a href="{link}">{title}</a></span></div>`
    if feedIndexPattern != feedTar.IndexPattern {
        t.Fatalf("%s: failed to parse index pattern, expected %s, got %s", config_file, feedIndexPattern, feedTar.IndexPattern)
    }

    feedContentPattern := `<div class="clearfix visible-xs niu2-clearfix"></div>{content}<div id="content-comments">`
    if feedContentPattern != feedTar.ContentPattern {
        t.Fatalf("%s: failed to parse content pattern, expected %s, got %s", config_file, feedContentPattern, feedTar.ContentPattern)
    }

    feedPath := `blog.atime.me.xml`
    if feedPath != feedTar.Path {
        t.Fatalf("%s: failed to parse path, expected %s, got %s", config_file, feedPath, feedTar.Path)
    }
}

func TestUnifyURL(t *testing.T) {
    rawURL := "atime.me"
    unifiedURL := "http://atime.me"

    if uniurl := unifyURL(rawURL); uniurl != unifiedURL {
        t.Fatalf("failed to unify raw url %s", rawURL)
    }

    rawURL = "http://atime.me"
    if uniurl := unifyURL(rawURL); uniurl != unifiedURL {
        t.Fatalf("failed to unify raw url %s", rawURL)
    }
}

/*
func TestCrawlHtml(t *testing.T) {
    url := "blog.atime.me/agreement.html"
    cache, err := CrawlHtml(url)
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
    conf, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatalf("failed to parse example_config.json")
    }

    if _, err = os.Stat(conf.CacheDB); nil == err {
        os.Remove(conf.CacheDB)
    }

    CreateDbScheme(conf.CacheDB)

    url := "blog.atime.me/agreement.html"
    cache, err := FetchHtml(url, conf.CacheDB)
    if nil != err {
        t.Fatalf("failed to fetch html %s", err)
    }

    if url != cache.URL {
        t.Fatalf("wrong html cache, url not match")
    }

    cache2, err := GetHtmlCacheByURL(url, conf.CacheDB)
    if nil != err {
        t.Fatalf("html cache not saved for url %s", url)
    }

    if cache.URL != cache2.URL ||
        cache.LastModify != cache2.LastModify ||
        bytes.Compare(cache.Html, cache2.Html) {

        t.Fatalf("html cache not match")
    }

    os.Remove(conf.CacheDB)
}

func TestCheckPatterns(t *testing.T) {
    invalidTargets := [...]Target {
        Target{ IndexPattern: "", ContentPattern: "" },
        Target{ IndexPattern: "abc", ContentPattern: "" },
        Target{ IndexPattern: "abc", ContentPattern: "cde" },
        Target{ IndexPattern: "abc{link}", ContentPattern: "cde" },
        Target{ IndexPattern: "{title} {link}abc{title}", ContentPattern: "{content}" },
        Target{ IndexPattern: "{*}abc{title}", ContentPattern: "{title}" },
        Target{ IndexPattern: "{link}abc{title}", ContentPattern: "{title}{content}" },
        Target{ IndexPattern: "{link}abc{title}", ContentPattern: "{link}{*}{content}" },
    }

    validTargets := [...]Target {
        Target{ IndexPattern: "{link}abc{title}", ContentPattern: "{*}{content}" },
        Target{ IndexPattern: "{link}abc{*}cde{title}", ContentPattern: "{content}" },
    }

    for _, tar := range invalidTargets {
        if CheckPatterns(&tar) {
            t.Fatal("check patterns failed: IndexPattern %s, ContentPattern %s", tar.IndexPattern, tar.ContentPattern)
        }
    }

    for _, tar := range validTargets {
        if !CheckPatterns(&tar) {
            t.Fatal("check pattern failed: IndexPattern %s, ContentPattern %s", tar.IndexPattern, tar.ContentPattern)
        }
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

func TestFilterHtmlWithoutPattern(t *testing.T) {
    conf, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatal("failed to parse example_config.json")
    }

    tar := conf.Targets[0]
    htmlData, err := Crawl(tar.URL)
    if nil != err {
        t.Fatal("failed to download web page")
    }
    htmlData = MinifyHtml(htmlData)

    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        t.Fatalf("filter without index pattern failed for target %s", tar.URL)
    }
}

func TestDB(t *testing.T) {
    conf, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatal("failed to parse example_config.json")
    }

    if _, err = os.Stat(conf.CacheDB); nil == err {
        os.Remove(conf.CacheDB)
    }

    err = CreateDbScheme(conf.CacheDB)
    if nil != err {
        t.Fatalf("failed to create db %s: %s", conf.CacheDB, err)
    }

    cache := []HtmlCache {
        HtmlCache { URL: "blog.atime.me", LastModify: "Mon, 25 Nov 2013 19:43:31 GMT", Html: []byte("hello world") },
        HtmlCache { URL: "atime.me", LastModify: "Mon, 25 Nov 2013 16:43:31 GMT", Html: []byte("hello world") } }

    err = PutHtmlCache(conf.CacheDB, cache)
    if nil != err {
        t.Fatalf("failed to insert records to db %s: %s", conf.CacheDB, err)
    }

    cache2, err := GetHtmlCacheByURL(conf.CacheDB, "no.cache")
    if nil == err {
        t.Fatalf("should not get html cache from an non-exist url")
    }

    cache2, err = GetHtmlCacheByURL(conf.CacheDB, cache[0].URL)
    if nil != err {
        t.Fatalf("failed to get html cache for url %s", cache[0].URL)
    }
    if cache2.URL != cache[0].URL ||
        cache2.LastModify != cache[0].LastModify ||
        0 != bytes.Compare(cache2.Html, cache[0].Html) {

        t.Fatalf("got wrong html cache")
    }

    os.Remove(conf.CacheDB)
}

func TestParseIndexAndContentHtml(t *testing.T) {
    conf, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatal("failed to parse example_config.json")
    }

    for _, tar := range conf.Targets {
        entries, ok := ParseIndexHtml(tar)
        if !ok {
            t.Fatalf("failed to parse index html %s", tar.URL)
        }

        for _, entry := range entries {
            if !ParseContentHtml(tar, &entry) {
                t.Fatalf("failed to parse content html %s", tar.URL)
            }

            println("title", entry.Title)
            println("link", entry.Link)
            println("content length", len(entry.Content))
            println("content summary", string(entry.Content)[:200])
            println("=====\n")
        }
    }
}

