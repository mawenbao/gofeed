package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sync"
)

func init() {
	//log.SetFlags(log.Lshortfile)
}

var (
	gVerbose = flag.Bool("v", false, "be verbose")
	gDebug   = flag.Bool("d", false, "debug mode")
	gCPUNum  = flag.Int("c", runtime.NumCPU(), "number of cpus to run simultaneously")
)

func showUsage() {
	fmt.Printf("Usage %s [-v][-d][-c cpu_number] json_config_file\n\n", os.Args[0])
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

	// check flag values
	if *gCPUNum > runtime.NumCPU() {
		log.Printf("[WARN] cpu number %d too big, wil be set to actual number of your cpus: %d", *gCPUNum, runtime.NumCPU())
		*gCPUNum = runtime.NumCPU()
	}

	// debug mode is verbose
	if *gDebug {
		*gVerbose = true
	}

	// parse json configuration first
	feedTargets := ParseJsonConfig(args[0])
	cacheDB := feedTargets[0].CacheDB

	var err error

	// create cache db if not exists
	if _, err = os.Stat(cacheDB); nil != err && os.IsNotExist(err) {
		log.Printf("creating cache database %s", cacheDB)
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
				log.Printf("[ERROR] failed to generate rss %s", feedTar.FeedPath)
			} else {
				err = ioutil.WriteFile(feedTar.FeedPath, rss2FeedStr, 0644)
				if nil != err {
					log.Printf("[ERROR] failed to save feed at %s", feedTar.FeedPath)
				}
				if *gVerbose {
					log.Printf("[DONE] saving feed at %s", feedTar.FeedPath)
				}
			}
		}(feedTar)
	}

	wg.Wait()
}
