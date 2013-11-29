package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

// exits on any parse or check error
func ParseJsonConfig(path string) (feedTargets []*FeedTarget) {
	var conf Config
	if configData, err := ioutil.ReadFile(path); nil != err {
		log.Fatalf("%s: %s", path, err)
	} else {
		err = json.Unmarshal(configData, &conf)
		if nil != err {
			log.Fatalf("%s: %s", path, err)
		}
		if 0 == len(conf.Targets) {
			log.Fatalf("[ERROR] no targets in config file %s", path)
		}
	}

	// check cache db
	if "" == conf.CacheDB {
		conf.CacheDB = "cache.db"
	}
	var err error
	conf.CacheDB, err = filepath.Abs(conf.CacheDB)
	if nil != err {
		log.Fatalf("failed to abs CacheDB %s", conf.CacheDB)
		return
	}

	feedTargets = make([]*FeedTarget, len(conf.Targets))

	// check target settings
	for i := 0; i < len(conf.Targets); i++ {
		tar := &conf.Targets[i]
		feedTar := &FeedTarget{
			CacheDB:     conf.CacheDB,
			ReqInterval: tar.ReqInterval,
			Description: tar.Description,
		}

		// abs feed path
		tar.FeedPath, err = filepath.Abs(tar.FeedPath)
		if nil != err {
			log.Fatalf("error abs feed path %s", tar.FeedPath)
		}
		// check feed existense
		if feedFile, err := os.Stat(tar.FeedPath); nil == err {
			if feedFile.IsDir() {
				log.Fatalf("config error, feed path is a directory %s", tar.FeedPath)
			} else {
				log.Printf("[WARN] feed %s already exists, will overwrite it", tar.FeedPath)
			}
		}
		// check parent folder existense
		absParentDir := filepath.Dir(tar.FeedPath)
		feedParentDir, err := os.Stat(absParentDir)
		if nil == err && !feedParentDir.IsDir() {
			log.Fatalf("parent directory of feed %s is actually a FILE %s", tar.FeedPath, absParentDir)
		}
		if nil != err {
			log.Printf("parent directory %s of feed %s does not exist, will create it", absParentDir, tar.FeedPath)
			err = os.MkdirAll(absParentDir, 0755)
			if nil != err {
				log.Fatalf("failed to create directory %s for feed %s", absParentDir, tar.FeedPath)
			}
		}

		feedTar.FeedPath = tar.FeedPath

		// set feed title
		if "" != tar.Title {
			feedTar.Title = tar.Title
		} else {
			feedTar.Title = filepath.Base(tar.FeedPath)
		}

		// check patterns
		if !CheckPatterns(tar) {
			log.Fatal("error parsing configuration: empty index/content pattern")
		}

		// compile patterns
		err = CompileIndexContentPatterns(feedTar, tar)
		if nil != err {
			log.Fatalf("failed to compile index/content patterns for feed target %s: %s", feedTar.FeedPath, err)
		}

		// normalize url
		if 0 == len(tar.URLs) {
			log.Fatalf("no urls for %s", tar.FeedPath)
		}
		feedTar.URLs = make([]*url.URL, len(tar.URLs))
		for urlInd, rawURL := range tar.URLs {
			normalURL, err := url.Parse(NormalizeURLStr(rawURL))
			if nil != err {
				log.Fatalf("error parsing target url %s: %s", rawURL, err)
			}
			feedTar.URLs[urlInd] = normalURL
		}

		// add feed target
		feedTargets = append(feedTargets[:i], feedTar)
	}

	return
}
