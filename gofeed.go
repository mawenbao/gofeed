package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"sort"
	"sync"
)

var (
	gVerbose           = flag.Bool("v", false, "be verbose")
	gDebug             = flag.Bool("d", false, "debug mode")
	gCPUNum            = flag.Int("c", runtime.NumCPU(), "number of cpus to run simultaneously")
	gLogfile           = flag.String("l", "", "path of the log file")
	gAlwaysUseCache    = flag.Bool("a", false, "use cache if failed to download web page")
	gGzipCompressLevel = flag.Int("z", 9, "compression level when saving html cache with gzip in the cache database.\n\t0-9 acceptable where 0 means no compression")
	gVersion           = flag.Bool("version", false, "print gofeed version")
)

func showUsage() {
	fmt.Printf("Usage %s [-version][-v][-d][-c cpu_number][-l log_file][-k][-z compression_level] json_config_file\n\n", os.Args[0])
	fmt.Printf("Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = showUsage
	flag.Parse()

	// print gofeed version
	if *gVersion {
		fmt.Printf("%s %s by mawenbao\n", GOFEED_NAME, GOFEED_VERSION)
		return
	}

	args := flag.Args()
	if 0 == len(args) || 1 < len(args) {
		flag.Usage()
		return
	}

	// check flag values
	if *gCPUNum > runtime.NumCPU() {
		log.Printf("[WARN] cpu number %d too big, wil be set to actual number of your cpus: %d", *gCPUNum, runtime.NumCPU())
		*gCPUNum = runtime.NumCPU()
		runtime.GOMAXPROCS(*gCPUNum)
	}

	if *gGzipCompressLevel < 0 || *gGzipCompressLevel > 9 {
		log.Printf("[WARN] gzip compression level invalid: %d, will use level 9 to compress html cache data", *gGzipCompressLevel)
		*gGzipCompressLevel = 9
	}

	var logfile *os.File
	var err error

	if "" != *gLogfile {
		logfile, err = os.OpenFile(*gLogfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if nil != err {
			log.Fatalf("failed to open/create logfile %s: %s", *gLogfile, err)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	}

	if *gDebug {
		// debug mode is verbose
		*gVerbose = true
		// print file name and line number too
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	}

	// parse json configuration first
	feedTargets := ParseJsonConfig(args[0])
	cacheDB := feedTargets[0].CacheDB

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
			log.Printf("start processing target %s", feedTar.FeedPath)
			// parse feed entry title and link
			feed, ok := ParseIndexHtml(feedTar)
			if !ok {
				log.Printf("[ERROR] failed to parse feed target %s", feedTar.FeedPath)
			}

			// remove duplicate entries(by link)
			RemoveDuplicatEntries(feed)

			// parse feed entry description
			if !ParseContentHtml(feedTar, feed) {
				log.Printf("[ERROR] failed to parse content html for feed target %s", feedTar.FeedPath)
			}

			// fill empty pubdates
			SetPubDates(feed)

			// remove junk code in entry content
			RemoveJunkContent(feed)

			// sort feed entries on pubdatet desc
			sort.Sort(sort.Reverse(FeedEntriesSortByPubDate(feed.Entries)))

			// generate rss2 feed
			rss2FeedStr, err := GenerateRss2Feed(feed)
			if nil != err {
				log.Printf("[ERROR] failed to generate rss %s", feedTar.FeedPath)
			} else {
				err = ioutil.WriteFile(feedTar.FeedPath, rss2FeedStr, 0644)
				if nil != err {
					log.Printf("[ERROR] failed to save feed at %s: %s", feedTar.FeedPath, err)
				} else {
					log.Printf("[DONE] saving feed at %s", feedTar.FeedPath)
				}
			}
		}(feedTar)
	}

	wg.Wait()
}
