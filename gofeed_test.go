package main

import(
    "testing"
    "io/ioutil"
    "bytes"
)

func TestParseJsonConfig(t *testing.T) {
    targets, err := ParseJsonConfig("example_config.json")
    config_file := "example_config.json"

    if nil != err || 1 != len(targets.Targets) {
        t.Fatalf("Failed to parse json config file %s: %s\n", config_file, err)
    }

    feedTar := targets.Targets[0]

    feedURL := `blog.atime.me`
    if feedURL != feedTar.URL {
        t.Fatalf("%s: failed to parse url, expected %s, got %s\n", config_file, feedURL, feedTar.URL)
    }

    feedIndexPattern := `<div class="niu2-index-article-title"><span><a href="{link}">{title}</a></span></div>`
    if feedIndexPattern != feedTar.IndexPattern {
        t.Fatalf("%s: failed to parse index pattern, expected %s, got %s\n", config_file, feedIndexPattern, feedTar.IndexPattern)
    }

    feedContentPattern := `<div class="niu2-lastmod-box">{*}</div>{content}<div id="content-comments">`
    if feedContentPattern != feedTar.ContentPattern {
        t.Fatalf("%s: failed to parse content pattern, expected %s, got %s\n", config_file, feedContentPattern, feedTar.ContentPattern)
    }

    feedPath := `blog.atime.me.xml`
    if feedPath != feedTar.Path {
        t.Fatalf("%s: failed to parse path, expected %s, got %s\n", config_file, feedPath, feedTar.Path)
    }
}

func TestUnifyURL(t *testing.T) {
    rawURL := "atime.me"
    unifiedURL := "http://atime.me"

    if uniurl, err := unifyURL(rawURL); err != nil || uniurl != unifiedURL {
        t.Fatalf("failed to unify raw url %s", rawURL)
    }
}

func TestCrawl(t *testing.T) {
    url := "blog.atime.me/agreement.html"
    data, err := Crawl(url)
    if nil != err {
        t.Fatalf("failed to crawl %s: %s\n", url, err)
    }

    testFile := "test_data/test_crawl.html"
    expectData, err := ioutil.ReadFile(testFile)

    if nil != err {
        t.Fatalf("failed to read %s: %s\n", testFile, err)
    }

    if 0 != bytes.Compare(expectData, data) {
        t.Fatalf("html data crawled from %s not equal to %s", url, testFile)
    }

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
        if tar.CheckPatterns() {
            t.Fatal("check patterns failed: IndexPattern %s, ContentPattern %s", tar.IndexPattern, tar.ContentPattern)
        }
    }

    for _, tar := range validTargets {
        if !tar.CheckPatterns() {
            t.Fatal("check pattern failed: IndexPattern %s, ContentPattern %s", tar.IndexPattern, tar.ContentPattern)
        }
    }
}

