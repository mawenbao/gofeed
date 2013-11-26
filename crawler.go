package main

import(
    "net/http"
    "log"
    "io/ioutil"
    "strings"
    "time"
)

// download html
func RequestHtml(cache *HtmlCache) (err error) {
    if *gVerbose {
        log.Printf("trying to download web page %s", cache.URL)
    }

    unifiedURL := unifyURL(cache.URL)

    if *gVerbose && cache.URL != unifiedURL {
        log.Printf("url %s unified to %s", cache.URL, unifiedURL)
    }

    req, err := http.NewRequest("GET", unifiedURL, nil)
    if nil != err {
        log.Printf("failed to create http request for %s: %s", unifiedURL, err)
        return
    }

    req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:12.0) Gecko/20100101 Firefox/21.0")
    if "" != cache.CacheControl {
        req.Header.Set("Cache-Control", cache.CacheControl)
    }
    if "" != cache.LastModified {
        req.Header.Set("If-Modified-Since", cache.LastModified)
    }
    if "" != cache.Etag {
        req.Header.Set("If-None-Match", cache.Etag)
    }

    client := new(http.Client)
    resp, err := client.Do(req)
    if nil != err {
        log.Printf("failed to fetch %s: %s", cache.URL, err)
        return
    }
    defer resp.Body.Close()

    if http.StatusNotModified == resp.StatusCode {
        // not modified, use cache
        cache.Status = CACHE_NOT_MODIFIED
        if *gVerbose {
            log.Printf("%s not modified", cache.URL)
        }
        return
    } else {
        if CACHE_NEW != cache.Status {
            cache.Status = CACHE_MODIFIED
            if *gVerbose {
                log.Printf("cache of %s has been modified", cache.URL)
            }
        }
        cache.Html, err = ioutil.ReadAll(resp.Body)
        if nil != err {
            log.Printf("failed to read response body for %s: %s", cache.URL, err)
            return
        }

        if cacheCtl, ok := resp.Header["Cache-Control"]; ok {
            cache.CacheControl = cacheCtl[0]
        } else {
            cache.CacheControl = ""
        }
        if lastmod, ok := resp.Header["Last-Modified"]; ok {
            cache.LastModified = lastmod[0]
        } else {
            cache.LastModified = ""
        }
        if etag, ok := resp.Header["Etag"]; ok {
            cache.Etag = etag[0]
        } else {
            cache.Etag = ""
        }
    }

    return
}

// try to retrive html from cache first
func FetchHtml(rawURL, dbPath string) (cache HtmlCache, err error) {
   cache, err = GetHtmlCacheByURL(dbPath, rawURL)

   if nil != err {
       // cache not found
       cache = HtmlCache { Status: CACHE_NEW }
   } else if "" != cache.Expires {
       var expireDate time.Time
       expireDate, err = http.ParseTime(cache.Expires)
       if nil != err {
           log.Printf("failed to parse expire date %s: %s", cache.Expires, err)
           return
       }
       if !time.Now().Before(expireDate) {
           // cache has expired
           cache.Etag = ""
           cache.LastModified = ""
           cache.CacheControl = ""
           if *gVerbose {
               log.Printf("cache of %s has expired", cache.URL)
           }
       }
   }

   cache.URL = rawURL
   err = RequestHtml(&cache)
   if nil != err {
       if CACHE_NEW == cache.Status {
           log.Printf("failed to download web page %s, just ignore it", rawURL)
           return
       } else {
           log.Printf("failed to download web page %s, use cache instead", rawURL)
       }
   }

   // extract html title
   cache.Title = ExtractHtmlTitle(cache.Html)
   if "" == cache.Title {
       log.Printf("failed to extract html title for %s", cache.URL)
   }

   switch cache.Status {
   case CACHE_NEW:
       // save html cache
       PutHtmlCache(dbPath, []HtmlCache { cache })
   case CACHE_MODIFIED:
       // update html cache
       UpdateHtmlCache(dbPath, []HtmlCache { cache })
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

