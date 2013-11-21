package main

import (
    "net/http"
    "log"
    "io/ioutil"
    "strings"
)

func Crawl(rawURL string) (data []byte, err error) {
    unifiedURL, err := unifyURL(rawURL)
    if nil != err {
        log.Printf("failed to unify raw url %s\n", rawURL)
        return
    }

    resp, err := http.Get(unifiedURL)
    if nil != err {
        log.Printf("Failed to fetch %s: %s\n", rawURL, err)
        return
    }
    defer resp.Body.Close()

    data, err = ioutil.ReadAll(resp.Body)
    if nil != err {
        log.Printf("Failed to read response body for %s: %s\n", rawURL, err)
        return
    }

    return
}

func unifyURL(rawURL string) (unifiedURL string, err error) {
    if !strings.HasPrefix(rawURL, HTTP_SCHEME) {
        unifiedURL = HTTP_SCHEME + SCHEME_SUFFIX + rawURL
    }

    return
}

