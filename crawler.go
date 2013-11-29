package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// download html
func SendHttpRequest(cache *HtmlCache) (resp *http.Response, err error) {
	if *gVerbose {
		log.Printf("start to request %s", cache.URL)
	}

	req, err := http.NewRequest("GET", cache.URL.String(), nil)
	if nil != err {
		log.Printf("[ERROR] failed to create http request for %s: %s", cache.URL, err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:12.0) Gecko/20100101 Firefox/21.0")
	// set cache related headers
    if "" != cache.CacheControl {
        req.Header.Set("Cache-Control", cache.CacheControl)
    }
    if nil != cache.LastModified {
        req.Header.Set("If-Modified-Since", cache.LastModified.Format(http.TimeFormat))
    }
    if "" != cache.Etag {
        req.Header.Set("If-None-Match", cache.Etag)
    }

    // set cache date as request date
    dateNow := time.Now()
    cache.Date = &dateNow

    // send request
	client := new(http.Client)
	resp, err = client.Do(req)
	if nil != err {
		log.Printf("[ERROR] http client failed to send request to %s: %s", cache.URL.String(), err)
		return
	}

	return
}

// will close response body
func ParseHttpResponse(resp *http.Response, cache *HtmlCache) (err error) {
	defer resp.Body.Close()

    // set cache date
    if cacheDate, ok := resp.Header["Date"]; ok {
        cache.Date = new(time.Time)
        *cache.Date, err = http.ParseTime(cacheDate[0])
        if nil != err {
            log.Printf("[ERROR] failed to parse http response Date header %s: %s", cacheDate, err)
        }
    }

    if http.StatusNotModified == resp.StatusCode {
		// not modified, use cache
		cache.Status = CACHE_NOT_MODIFIED
		if *gVerbose {
			log.Printf("cache for %s not modified", cache.URL.String())
		}
		return
	} else {
		// change status of expired cache to modified
		if CACHE_NEW != cache.Status {
			cache.Status = CACHE_MODIFIED
			if *gVerbose {
				log.Printf("cache for %s has been modified", cache.URL.String())
			}
		}
		cache.Html, err = ioutil.ReadAll(resp.Body)
		if nil != err {
			log.Printf("[ERROR] failed to read response body for %s: %s", cache.URL.String(), err)
			return
		}

		if cacheCtl, ok := resp.Header["Cache-Control"]; ok {
			cache.CacheControl = cacheCtl[0]
		} else {
			cache.CacheControl = ""
		}
		if lastmod, ok := resp.Header["Last-Modified"]; ok {
            cache.LastModified = new(time.Time)
			*cache.LastModified, err = http.ParseTime(lastmod[0])
			if nil != err {
				log.Printf("[ERROR] error parsing http Last-Modified response header %s: %s", lastmod[0], err)
			}
		}
		if expireStr, ok := resp.Header["Expires"]; ok {
            cache.Expires = new(time.Time)
			*cache.Expires, err = http.ParseTime(expireStr[0])
			if nil != err {
				log.Printf("[ERROR] error parsing http Expires response header %s: %s", expireStr[0], err)
			}
		}
		if etag, ok := resp.Header["Etag"]; ok {
			cache.Etag = etag[0]
		} else {
			cache.Etag = ""
		}
	}

	return
}

func FetchHtml(normalURL *url.URL, dbPath string) (cache *HtmlCache, err error) {
	// try to retrive html from cache first
	cache, err = GetHtmlCacheByURL(dbPath, normalURL.String())

	if nil == cache || nil != err {
		// cache not found
		cache = &HtmlCache{Status: CACHE_NEW}
	} else {
        // cache not expired, reuse it
        if time.Now().Before(cache.Date.Add(time.Second * ExtractMaxAge(cache.CacheControl))) ||
        (nil != cache.Expires && time.Now().After(*cache.Expires)) {
            log.Printf("cache for %s has not expired", cache.URL.String())
            return cache, nil
		}
	}

    // cache not found or expired, send new request
	cache.URL = normalURL
	resp, err := SendHttpRequest(cache)
	if nil != err {
		log.Printf("[ERROR] failed sending http request to %s: %s", cache.URL.String(), err)
		// stop
		return
	}

	err = ParseHttpResponse(resp, cache)
	if nil != err {
		if CACHE_NEW == cache.Status {
			log.Printf("[ERROR] failed to download web page %s, just ignore it", normalURL.String())
			return
		} else {
			log.Printf("[ERROR] failed to download web page %s, use cache instead", normalURL.String())
		}
	}

	switch cache.Status {
	case CACHE_NEW:
		// save html cache
		PutHtmlCache(dbPath, []*HtmlCache{cache})
	case CACHE_MODIFIED:
		// update html cache
		UpdateHtmlCache(dbPath, []*HtmlCache{cache})
	}

	return
}
