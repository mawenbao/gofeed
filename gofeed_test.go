package main

import(
    "testing"
    _ "io/ioutil"
    _ "bytes"
)

func TestParseJsonConfig(t *testing.T) {
    targets, err := ParseJsonConfig("example_config.json")
    config_file := "example_config.json"

    if nil != err {
        t.Fatalf("Failed to parse json config file %s: %s", config_file, err)
    }

    feedTar := targets.Targets[0]

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
func TestCrawl(t *testing.T) {
    url := "blog.atime.me/agreement.html"
    data, err := Crawl(url)
    if nil != err {
        t.Fatalf("failed to crawl %s: %s", url, err)
    }

    testFile := "test_data/test_crawl.html"
    expectData, err := ioutil.ReadFile(testFile)

    if nil != err {
        t.Fatalf("failed to read %s: %s", testFile, err)
    }

    if 0 != bytes.Compare(expectData, data) {
        t.Fatalf("html data crawled from %s not equal to %s", url, testFile)
    }

}
*/

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
    targets, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatal("failed to parse example_config.json")
    }

    tar := targets.Targets[0]
    htmlData, err := Crawl(tar.URL)
    if nil != err {
        t.Fatal("failed to download web page")
    }
    htmlData = MinifyHtml(htmlData)

    if !FilterHtmlWithoutPattern(htmlData, tar.IndexPattern) {
        t.Fatalf("filter without index pattern failed for target %s", tar.URL)
    }
}

func TestParseIndexAndContentHtml(t *testing.T) {
    targets, err := ParseJsonConfig("example_config.json")
    if nil != err {
        t.Fatal("failed to parse example_config.json")
    }

    for _, tar := range targets.Targets {
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

