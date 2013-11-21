package main

import(
    "testing"
)

func TestParseJsonConfig(t *testing.T) {
    targets := ParseJsonConfig("example_config.json")
    config_file := "example_config.json"

    if 1 != len(targets.Targets) {
        t.Fatal("Failed to parse json config file", config_file)
    }

    feedTar := targets.Targets[0]

    feedURL := `http://blog.atime.me`
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

