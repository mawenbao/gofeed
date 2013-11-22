package main

import(
    "log"
    "strings"
)

const(
    SCHEME_SUFFIX = "://"
    HTTP_SCHEME = "http"

    TITLE_NAME = "title"
    PATTERN_TITLE = "{" + TITLE_NAME + "}"
    PATTERN_TITLE_REG = "(?P<" + TITLE_NAME + ">.*)"

    LINK_NAME = "link"
    PATTERN_LINK = "{" + LINK_NAME + "}"
    PATTERN_LINK_REG = "(?P<" + LINK_NAME + ">.*)"

    CONTENT_NAME = "content"
    PATTERN_CONTENT = "{" + CONTENT_NAME + "}"
    PATTERN_CONTENT_REG = "(?P<" + CONTENT_NAME + ">)"

    PATTERN_ANY = "{any}"
    PATTERN_ANY_REG = ".*"
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
        PATTERN_ANY, PATTERN_ANY_REG,
        PATTERN_TITLE, PATTERN_TITLE_REG,
        PATTERN_LINK, PATTERN_LINK_REG,
        PATTERN_CONTENT, PATTERN_CONTENT_REG)

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

