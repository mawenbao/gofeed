package main

import(
    "os"
    "fmt"
    "flag"
    "log"
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
                log.Printf("[ERROR] failed to parse content html %s", tar.URL)
            }
        }

        // generate rss2 feed
        feedStr, err := GenerateRss2Feed(indexCache, entries)
        if nil != err {
            log.Fatalf("[ERROR] failed to generate rss")
        }

        if *gVerbose {
            log.Printf("saving feed at %s", tar.FeedPath)
        }
        err = ioutil.WriteFile(tar.FeedPath, feedStr, 0644)
        if nil != err {
            log.Fatal("[ERROR] failed to save feed at %s", tar.FeedPath)
        }
    }
}

