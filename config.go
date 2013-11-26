package main

import(
    "io/ioutil"
    "log"
    "encoding/json"
    "path/filepath"
    "net/url"
)

func ParseJsonConfig(path string) (conf Config, err error) {
    if configData, err := ioutil.ReadFile(path); nil != err {
        log.Fatalf("%s: %s", path, err)
    } else {
        err = json.Unmarshal(configData, &conf)
        if nil != err {
            log.Fatalf("%s: %s", path, err)
        }
    }

    // check cache db
    if "" == conf.CacheDB {
        conf.CacheDB = "cache.db"
    }
    conf.CacheDB, err = filepath.Abs(conf.CacheDB)
    if nil != err {
        log.Fatalf("failed to abs CacheDB %s", conf.CacheDB)
        return
    }

    // check target settings
    for i := 0; i < len(conf.Targets); i++ {
        tar := &conf.Targets[i]
        // abs feed path
        tar.FeedPath, err = filepath.Abs(tar.FeedPath)
        if nil != err {
            log.Fatalf("error abs feed path %s for target %s", tar.FeedPath, tar.URL)
        }
        // index pattern and content pattern must not be empty
        if "" == tar.IndexPattern || "" == tar.ContentPattern {
            log.Fatal("error parsing configuration: empty index/content pattern")
        }
        // normalize url
        normalURL, err := url.Parse(NormalizeURLStr(tar.URL))
        if nil != err {
            log.Fatalf("error parsing target url %s: %s", tar.URL, err)
        }
        tar.URL = normalURL.String()
    }
    return
}

