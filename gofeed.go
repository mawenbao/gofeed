package main

import(
    "os"
    "fmt"
    "flag"
    "log"
    "io/ioutil"
    "sync"
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
    feedTargets := ParseJsonConfig(args[0])
    cacheDB := feedTargets[0].CacheDB

    var err error

    // create cache db if not exists
    if *gVerbose {
        log.Printf("Creating cache database %s", cacheDB)
    }

    if _, err = os.Stat(cacheDB); nil != err && os.IsNotExist(err) {
        err = CreateDBScheme(cacheDB)
        if nil != err {
            log.Fatalf("[ERROR] failed to create cache database %s", cacheDB)
        }
    } else {
        log.Printf("found cache database %s", cacheDB)
    }

    var wg sync.WaitGroup

    for _, feedTar := range feedTargets {
        wg.Add(1)

        go func(feedTar *FeedTarget) {
            defer wg.Done()
            // parse feed entry title and link
            feed, ok := ParseIndexHtml(feedTar)
            if !ok {
                log.Printf("[ERROR] failed to parse feed target %s", feedTar.FeedPath)
            }

            // parse feed entry description
            if !ParseContentHtml(feedTar, feed) {
                log.Printf("[ERROR] failed to parse content html for feed target %s", feedTar.FeedPath)
            }

            // generate rss2 feed
            rss2FeedStr, err := GenerateRss2Feed(feed)
            if nil != err {
                log.Fatalf("[ERROR] failed to generate rss")
            }

            if *gVerbose {
                log.Printf("[DONE] saving feed at %s", feedTar.FeedPath)
            }
            err = ioutil.WriteFile(feedTar.FeedPath, rss2FeedStr, 0644)
            if nil != err {
                log.Printf("[ERROR] failed to save feed at %s", feedTar.FeedPath)
            }
        }(feedTar)
    }

    wg.Wait()
}

