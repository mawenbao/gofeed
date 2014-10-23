package main

import (
	"log"
	"regexp"
	"strings"
	"time"
)

func ExtractHtmlTitle(htmlData []byte) string {
	titleReg := regexp.MustCompile(HTML_TITLE_REG)
	matches := titleReg.FindSubmatch(htmlData)
	if len(matches) != 2 {
		return ""
	}
	return string(matches[1])
}

func FilterHtmlWithoutPattern(htmlData []byte, pattern string) bool {
	html := string(htmlData)
	for _, str := range PATTERN_ALL_REGEX.Split(pattern, -1) {
		if "" == str {
			continue
		}
		if !strings.Contains(html, str) {
			log.Printf("[ERROR] target html does not contain %s", str)
			if *gDebug {
				log.Println("======= debug: target html data =======")
				log.Println(string(htmlData))
				log.Println("==============")
			}

			return false
		}
	}

	return true
}

func ParseIndexHtml(feedTar *FeedTarget) (feed *Feed, ok bool) {
	feed = new(Feed)

	for urlInd, tarURL := range feedTar.URLs {
		// get cache
		indexCache, err := FetchHtml(tarURL, feedTar)
		if nil == indexCache || nil != err {
			log.Printf("[ERROR] failed to download index web page %s", tarURL.String())
			// just ignore the sucker, feed.URL = nil
			continue
		}

		// minify html
		htmlData := MinifyHtml(RemoveJunkContent(indexCache.Html))

		// extract feed entry title and link
		indRegs := FindIndexRegs(feedTar, tarURL)
		for ind, indexReg := range indRegs {
			if nil == indexReg {
				log.Printf("[ERROR] cannot find index regex for %s", tarURL.String())
				continue
			}

			// make a copy of source data
			var htmlDataCopy []byte
			if ind+1 == len(indRegs) {
				htmlDataCopy = htmlData
			} else {
				htmlDataCopy = make([]byte, len(htmlData))
				copy(htmlDataCopy, htmlData)
			}

			// filter html with index filter
			indexFilterReg := FindIndexFilterReg(feedTar, indexReg)
			if nil != indexFilterReg {
				htmlDataCopy = RegexpFilter(indexFilterReg, htmlDataCopy)
				if nil == htmlDataCopy {
					// failed to filter htmlData
					continue
				}
			}

			matches := indexReg.FindAllSubmatch(htmlDataCopy, -1)
			if nil == matches {
				log.Printf("[ERROR] failed to match index html %s, pattern %s did not match", tarURL.String(), indexReg.String())
				if *gDebug {
					log.Println("======= debug: target html data =======")
					log.Println(string(htmlDataCopy))
					log.Println("==============")
				}
				// ignore this
				continue
			}

			entries := make([]*FeedEntry, len(matches))
			for matchInd, match := range matches {
				entries[matchInd] = new(FeedEntry)
				entry := entries[matchInd] // pointer of FeedEntry
				entry.IndexPattern = indexReg
				for patInd, patName := range indexReg.SubexpNames() {
					switch patName {
					case PATTERN_TITLE:
						entry.Title = string(match[patInd])
					case PATTERN_LINK:
						// normalize entry link which may be relative
						entry.Link, err = tarURL.Parse(string(match[patInd]))
						if nil != err {
							log.Printf("[ERROR] error parsing entry link %s: %s", entry.Link, err)
						}
					case PATTERN_PUBDATE:
						var pubDate time.Time
						pubDate, err = ParsePubDate(FindPubDateReg(feedTar, feed.URL), string(match[patInd]))
						if nil != err {
							log.Printf("[ERROR] error parsing pubdate of link %s: %s", entry.Link, err)
						} else {
							entry.PubDate = &pubDate
						}
					}
				}
			}

			// add entries to feed
			feed.Entries = append(feed.Entries, entries...)
		}

		if 0 == urlInd {
			feed.Title = feedTar.Title
			feed.Description = feedTar.Description
			// use first index page and url
			feed.URL = tarURL
			dateNow := time.Now()
			if nil == indexCache.LastModified {
				feed.LastModified = &dateNow
			} else {
				feed.LastModified = indexCache.LastModified
			}
		} else {
			// use later lastmod time
			if nil != indexCache.LastModified && feed.LastModified.Before(*indexCache.LastModified) {
				feed.LastModified = indexCache.LastModified
			}
		}
	}
	return feed, true
}

func ParseContentHtml(feedTar *FeedTarget, feed *Feed) (ok bool) {
	validEntries := make([]*FeedEntry, 1)
	validEntryInd := 0
	for entryInd, entry := range feed.Entries {
		if nil == entry {
			log.Printf("[ERROR] failed to parse content html: entry is nil")
			continue
		}

		contentReg := FindContentReg(feedTar, feed.URL, entry.IndexPattern)
		if nil == contentReg {
			log.Printf("[ERROR] failed to find content regex for entry %s", entry.Link.String())
			return
		}

		// check entry link
		if nil == entry.Link {
			log.Printf("[ERROR] entry link is nil, ignore this. entry index is %d", entryInd)
			continue
		}

		// wait some time
		if *gVerbose {
			log.Printf("waiting for %d seconds before sending request to %s", feedTar.ReqInterval, entry.Link.String())
		}
		time.Sleep(feedTar.ReqInterval * time.Second)

		cache, err := FetchHtml(entry.Link, feedTar)
		if nil == cache || nil != err {
			log.Printf("[ERROR] failed to download web page %s, will remove this entry", entry.Link.String())
			// ignore this entry, entry.Cache = nil
			continue
		} else {
			validEntries = append(validEntries[:validEntryInd], entry)
			validEntryInd += 1
		}
		entry.Cache = cache

		htmlData := MinifyHtml(RemoveJunkContent(cache.Html))

		// filter html with content filter
		contFilterReg := FindContentFilterReg(feedTar, contentReg)
		if nil != contFilterReg {
			htmlData := RegexpFilter(contFilterReg, htmlData)
			if nil == htmlData {
				// failed to filter htmlData
				continue
			}
		}

		// extract feed entry content(description)
		match := contentReg.FindSubmatch(htmlData)
		if nil == match {
			log.Printf("[ERROR] failed to match content html %s, pattern %s match failed", entry.Link.String(), contentReg.String())
			if *gDebug {
				log.Println("======= debug: target html data =======")
				log.Println(string(htmlData))
				log.Println("==============")
			}
			// ignore this sucker
			continue
		}
		for patInd, patName := range contentReg.SubexpNames() {
			switch patName {
			case PATTERN_CONTENT:
				entry.Content = match[patInd]
			case PATTERN_PUBDATE:
				var pubDate time.Time
				pubDate, err = ParsePubDate(FindPubDateReg(feedTar, feed.URL), string(match[patInd]))
				if nil != err {
					log.Printf("[ERROR] error parsing pubdate of link %s: %s", entry.Link, err)
				} else {
					entry.PubDate = &pubDate
				}
			}
		}

		if 0 == len(entry.Content) {
			// just print a warning message if content is empty
			log.Printf("[WARN] feed entry has no description: %s", entry.Link.String())
		}
	}

	feed.Entries = validEntries

	return true
}
