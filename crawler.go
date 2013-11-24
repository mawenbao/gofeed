package main

import(
    "net/http"
    "log"
    "io/ioutil"
    "strings"
)

func Crawl(rawURL string) (data []byte, err error) {
    if *gVerbose {
        log.Printf("trying to download web page %s", rawURL)
    }

    unifiedURL := unifyURL(rawURL)

    if *gVerbose && rawURL != unifiedURL {
        log.Printf("url %s unified to %s", rawURL, unifiedURL)
    }

    resp, err := http.Get(unifiedURL)
    if nil != err {
        log.Printf("failed to fetch %s: %s", rawURL, err)
        return
    }
    defer resp.Body.Close()

    data, err = ioutil.ReadAll(resp.Body)
    if nil != err {
        log.Printf("failed to read response body for %s: %s", rawURL, err)
        return
    }

    return
}

func unifyURL(rawURL string) (unifiedURL string) {
    unifiedURL = rawURL

    if !strings.HasPrefix(rawURL, HTTP_SCHEME) {
        unifiedURL = HTTP_SCHEME + SCHEME_SUFFIX + rawURL
    }

    return
}

