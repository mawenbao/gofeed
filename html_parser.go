package main

import (
    "strings"
    "log"
    "regexp"
)

func FilterWithoutPattern(htmlData []byte, tar Target) (bool) {
    // @TODO should reuse the compile regex
    patterns := [...]string { PATTERN_ANY, PATTERN_CONTENT, PATTERN_LINK, PATTERN_TITLE }
    regStr := strings.Join(patterns, "|")
    reg, err := regexp.Compile(regStr)
    if nil != err {
        log.Printf("regular string %s cannot be compiled", regStr)
        return false
    }

    html := string(htmlData)
    for _, tarPat := range [...]string { tar.IndexPattern, tar.ContentPattern } {
        for _, str := reg.Split(tarPat) {
            if "" == str {
                continue
            }
            if !strings.Contains(html, str) {
                log.Printf("target %s does not contains %s", tar.URL, str)
                return false
            }
        }
    }

    return true
}

func ParseHtml(htmlData []byte, tar Target) (feeds []Feed) {
    html := string(htmlData)

    if !FilterWithoutPattern(htmlData, tar) {
        log.Printf("no match for target %s\n", tar.URL)
        return
    }

}

