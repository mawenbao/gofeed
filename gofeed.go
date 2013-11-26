package main

import(
    "os"
    "fmt"
    "flag"
    "log"
    "time"
    "net/http"
    "io/ioutil"
)

func init() {
    //log.SetFlags(log.Lshortfile)
}

var (
    gVerbose = flag.Bool("v", false, "be verbose")
)

func showUsage() {
    fmt.Printf("Usage %s [-v] json_config_file\n\n", os.Args[0])
    fmt.Printf("Flags:\n")
    flag.PrintDefaults()
}

func main() {
    flag.Usage = showUsage
    flag.Parse()

    args := flag.Args()
    if 0 == len(args) || 1 < len(args) {
        flag.Usage()
        return
    }

    // parse json configuration first
    conf, err := ParseJsonConfig(args[0])
    if nil != err {
        log.Fatalf("[ERROR] failed to parse json configuration %s", args[0])
    }

    // create cache db if not exists
    if *gVerbose {
        log.Printf("Creating cache database %s", conf.CacheDB)
    }

    if _, err = os.Stat(conf.CacheDB); nil != err && os.IsNotExist(err) {
        err = CreateDbScheme(conf.CacheDB)
        if nil != err {
            log.Fatalf("[ERROR] failed to create cache database %s", conf.CacheDB)
        }
    } else {
        log.Printf("found cache database %s", conf.CacheDB)
    }

    targetMap := make(map[string]*TargetGroup)

    for _, tar := range conf.Targets {
        // parse feed entry title and link
        indexCache, entries, ok := ParseIndexHtml(conf, tar)
        if !ok {
            log.Printf("[ERROR] failed to parse index html %s", tar.URL)
            continue
        }

        // parse feed entry description
        for i := 0; i < len(entries); i++ {
            if !ParseContentHtml(conf, tar, &entries[i]) {
                log.Printf("[ERROR] failed to parse content html %s", entries[i].Link)
            }
        }

        var tarGrp *TargetGroup
        if grp, ok := targetMap[tar.FeedPath]; ok {
            tarGrp = grp
            // use the latest LastModified
            tarGrp.IndexCache.LastModified, err = GetLaterTimeStr(tarGrp.IndexCache.LastModified, indexCache.LastModified)
            if nil != err {
                log.Printf("[ERROR] failed to get later time from %s and %s", tarGrp.IndexCache.LastModified, indexCache.LastModified)
                // on time parse error, use current time
                tarGrp.IndexCache.LastModified = time.Now().Format(http.TimeFormat)
            }
        } else {
            tarGrp = new(TargetGroup)
            // use the first target's index html cache
            tarGrp.FeedPath = tar.FeedPath
            tarGrp.IndexCache = indexCache
            targetMap[tar.FeedPath] = tarGrp
        }

        tarGrp.Targets = append(tarGrp.Targets, tar)
        tarGrp.Entries = append(tarGrp.Entries, entries...)
    }

    // generate rss2 feed
    for _, grp := range targetMap {
        feedStr, err := GenerateRss2Feed(grp.IndexCache, grp.Entries)
        if nil != err {
            log.Fatalf("[ERROR] failed to generate rss")
        }

        if *gVerbose {
            log.Printf("saving feed at %s", grp.FeedPath)
        }
        err = ioutil.WriteFile(grp.FeedPath, feedStr, 0644)
        if nil != err {
            log.Fatal("[ERROR] failed to save feed at %s", grp.FeedPath)
        }
    }
}

