package main

import(
    "net/http"
    "log"
    "io/ioutil"
    "strings"
)

// download web page
func CrawlHtml(rawURL string) (cache HtmlCache, err error) {
    if *gVerbose {
        log.Printf("trying to download web page %s", rawURL)
    }

    cache.URL = rawURL
    unifiedURL := unifyURL(rawURL)

    if *gVerbose && rawURL != unifiedURL {
        log.Printf("url %s unified to %s", rawURL, unifiedURL)
    }

    set If-Modified-Since in request header
    & User-agent

    resp, err := http.Get(unifiedURL)
    if nil != err {
        log.Printf("failed to fetch %s: %s", rawURL, err)
        return
    }
    defer resp.Body.Close()

    cache.LastModify =

    cache.Html, err = ioutil.ReadAll(resp.Body)
    if nil != err {
        log.Printf("failed to read response body for %s: %s", rawURL, err)
        return
    }

    return
}

// get from cache first
func FetchHtml(rawURL, dbPath string) (data []byte, err error) {
   cache, err := GetHtmlCacheByURL(dbPath, rawURL)

   // cache not found
   if nil != err {
       //if *gVerbose {
           log.Printf("cache for %s not found, download it now", rawURL)
       //}
       cache, err = Crawl(rawURL)
       if nil != err {
           log.Printf("failed to download web page %s", rawURL)
           return
       }

       // save html cache
       PutHtmlCache(dbPath, []HtmlCache { cache })
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

