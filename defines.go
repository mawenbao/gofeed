package main

import(
    "log"
    "strings"
)

const(
    SCHEME_SUFFIX = "://"
    HTTP_SCHEME = "http"

    PATTERN_TITLE = "{title}"
    PATTERN_LINK = "{link}"
    PATTERN_CONTENT = "{content}"
    PATTERN_ANY = "{any}"
)

type Target struct {
    URL string
    IndexPattern string
    ContentPattern string
    Path string
}

type TargetSlice struct {
    Targets []Target
}

type Feed struct {
    Title string
    Link string
    Content []byte
}

func PatternToRegex(pat string) string {
    r := strings.NewReplacer(
        PATTERN_ANY, ".*",
        PATTERN_TITLE, "",
        PATTERN_LINK, "",
        PATTERN_CONTENT, "")

    return r.Replace(pat)
}

func (tar *Target) CheckPatterns() bool {
    if nil == tar {
        log.Printf("invliad target, nil\n")
        return false
    }

    // IndexPattern should contain both {title} and {link}
    if "" == tar.IndexPattern {
        log.Print("index pattern is empty")
        return false
    }

    if 1 != strings.Count(tar.IndexPattern, PATTERN_TITLE) || 1 != strings.Count(tar.IndexPattern, PATTERN_LINK) {
        log.Printf("index pattern %s should contain 1 %s and 1 %s \n", tar.IndexPattern, PATTERN_TITLE, PATTERN_LINK)
        return false
    }

    // ContentPattern should contain {content} and should not contain {title} nor {link}
    if "" == tar.ContentPattern {
        log.Print("content pattern is empty")
        return false
    }

    if 1 != strings.Count(tar.ContentPattern, PATTERN_CONTENT) {
        log.Printf("content pattern %s should contain 1 %s\n", tar.ContentPattern, PATTERN_CONTENT)
        return false
    }

    if strings.Contains(tar.ContentPattern, PATTERN_TITLE) || strings.Contains(tar.ContentPattern, PATTERN_LINK) {
        log.Printf("%s should not contain %s or %s", tar.ContentPattern, PATTERN_TITLE, PATTERN_LINK)
        return false
    }

    return true
}

