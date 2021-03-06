package main

import (
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func PatternToRegex(pat string) string {
	r := strings.NewReplacer(
		PATTERN_ANY, PATTERN_ANY_REG,
		GenPDPName(PATTERN_TITLE), GenPDPRegexStr(PATTERN_TITLE, true, true),
		GenPDPName(PATTERN_LINK), GenPDPRegexStr(PATTERN_LINK, true, true),
		GenPDPName(PATTERN_CONTENT), GenPDPRegexStr(PATTERN_CONTENT, false, true),
		GenPDPName(PATTERN_FILTER), GenPDPRegexStr(PATTERN_FILTER, true, true),
		GenPDPName(PATTERN_PUBDATE), GenPDPRegexStr(PATTERN_PUBDATE, true, true),
		GenPDPName(PATTERN_YEAR), GenPDPRegexStr(PATTERN_YEAR, true, true),
		GenPDPName(PATTERN_MONTH), GenPDPRegexStr(PATTERN_MONTH, true, false),
		GenPDPName(PATTERN_DAY), GenPDPRegexStr(PATTERN_DAY, true, false),
		GenPDPName(PATTERN_HOUR), GenPDPRegexStr(PATTERN_HOUR, true, false),
		GenPDPName(PATTERN_MINUTE), GenPDPRegexStr(PATTERN_MINUTE, true, false),
		GenPDPName(PATTERN_SECOND), GenPDPRegexStr(PATTERN_SECOND, true, false),
	)

	return r.Replace(pat)
}

// IndexPattern must contain both {title} and {link}
// ContentPattern must contain {content}
// Either IndexPattern or ContentPattern may contain {pubdate}, but not both.
func CheckPatterns(tar *TargetConfig) bool {
	if nil == tar {
		log.Printf("[ERROR] invliad target, nil")
		return false
	}

	if (len(tar.URLs) != len(tar.IndexPatterns) && (1 != len(tar.IndexPatterns) && 1 != len(tar.URLs))) ||
		(len(tar.URLs) != len(tar.ContentPatterns) && (1 != len(tar.ContentPatterns) && 1 != len(tar.URLs))) {
		log.Printf("error parsing index/content patterns: len(URL) != len(IndexPattern|ContentPattern) || 1 != len(IndexPattern|ContentPattern")
		return false
	}

	for _, indexPat := range tar.IndexPatterns {
		// IndexPattern should contain both {title} and {link}
		if "" == indexPat {
			log.Print("[ERROR] index pattern is empty")
			return false
		}

		if 1 != strings.Count(indexPat, GenPDPName(PATTERN_TITLE)) || 1 != strings.Count(indexPat, GenPDPName(PATTERN_LINK)) {
			log.Printf("[ERROR] index pattern %s should contain 1 %s and 1 %s ", indexPat, GenPDPName(PATTERN_TITLE), GenPDPName(PATTERN_LINK))
			return false
		}
	}

	for _, contentPat := range tar.ContentPatterns {
		// ContentPattern should contain {content} and should not contain {title} nor {link}
		if "" == contentPat {
			log.Print("content pattern is empty")
			return false
		}

		if 1 != strings.Count(contentPat, GenPDPName(PATTERN_CONTENT)) {
			log.Printf("[ERROR] content pattern %s should contain 1 %s", contentPat, GenPDPName(PATTERN_CONTENT))
			return false
		}

		if strings.Contains(contentPat, GenPDPName(PATTERN_TITLE)) || strings.Contains(contentPat, GenPDPName(PATTERN_LINK)) {
			log.Printf("[ERROR] %s should not contain %s or %s", contentPat, GenPDPName(PATTERN_TITLE), GenPDPName(PATTERN_LINK))
			return false
		}
	}

	// check filter patterns
	if (0 != len(tar.IndexFilterPatterns) && len(tar.IndexFilterPatterns) != len(tar.IndexPatterns)) ||
		(0 != len(tar.ContentFilterPatterns) && len(tar.ContentFilterPatterns) != len(tar.ContentPatterns)) {
		log.Printf("error parsing filter patterns: length must be 0 or the same as Feed.IndexPattern/Feed.ContentPattern")
		return false
	}

	for _, indFilterPat := range tar.IndexFilterPatterns {
		if "" == indFilterPat {
			continue
		}
		if 1 > strings.Count(indFilterPat, GenPDPName(PATTERN_FILTER)) {
			log.Printf("[ERROR] index filter pattern %s should be empty or contain more than one %s", indFilterPat, GenPDPName(PATTERN_FILTER))
			return false
		}
	}

	for _, contFilterPat := range tar.ContentFilterPatterns {
		if "" == contFilterPat {
			continue
		}
		if 1 > strings.Count(contFilterPat, GenPDPName(PATTERN_FILTER)) {
			log.Printf("[ERROR] content filter pattern %s should be empty or contain more than one %s", contFilterPat, GenPDPName(PATTERN_FILTER))
			return false
		}
	}

	//@TODO check pubdate pattern

	return true
}

func CompilePatterns(feedTar *FeedTarget, tar *TargetConfig) (err error) {
	feedTar.IndexRegs = make([]*regexp.Regexp, len(tar.IndexPatterns))
	feedTar.ContentRegs = make([]*regexp.Regexp, len(tar.ContentPatterns))
	feedTar.IndexFilterRegs = make([]*regexp.Regexp, len(tar.IndexFilterPatterns))
	feedTar.ContentFilterRegs = make([]*regexp.Regexp, len(tar.ContentFilterPatterns))
	feedTar.PubDateRegs = make([]*regexp.Regexp, len(tar.PubDatePatterns))

	// index pattern
	for j := 0; j < len(tar.IndexPatterns); j++ {
		feedTar.IndexRegs[j], err = regexp.Compile(PatternToRegex(tar.IndexPatterns[j]))
		if nil != err {
			log.Printf("[ERROR] error compiling index pattern %s", tar.IndexPatterns[j])
			return
		}
	}

	// content pattern
	for j := 0; j < len(tar.ContentPatterns); j++ {
		feedTar.ContentRegs[j], err = regexp.Compile(PatternToRegex(tar.ContentPatterns[j]))
		if nil != err {
			log.Printf("[ERROR] error compiling content pattern %s", tar.ContentPatterns[j])
			return
		}
	}

	// index filter pattern
	for j := 0; j < len(tar.IndexFilterPatterns); j++ {
		if "" == strings.TrimSpace(tar.IndexPatterns[j]) {
			continue
		}
		feedTar.IndexFilterRegs[j], err = regexp.Compile(PatternToRegex(tar.IndexFilterPatterns[j]))
		if nil != err {
			log.Printf("[ERROR] error compiling index filter pattern %s", tar.IndexFilterPatterns[j])
			return
		}
	}

	// content filter pattern
	for j := 0; j < len(tar.ContentFilterPatterns); j++ {
		if "" == strings.TrimSpace(tar.ContentPatterns[j]) {
			continue
		}
		feedTar.ContentFilterRegs[j], err = regexp.Compile(PatternToRegex(tar.ContentFilterPatterns[j]))
		if nil != err {
			log.Printf("[ERROR] error compiling content filter pattern %s", tar.ContentFilterPatterns[j])
			return
		}
	}

	// publish date pattern
	for j := 0; j < len(tar.PubDatePatterns); j++ {
		if "" == strings.TrimSpace(tar.PubDatePatterns[j]) {
			continue
		}
		feedTar.PubDateRegs[j], err = regexp.Compile(PatternToRegex(tar.PubDatePatterns[j]))
		if nil != err {
			log.Printf("[ERROR] error compiling publish date pattern %s", tar.PubDatePatterns[j])
			return
		}
	}

	return
}

// return -1, cache lives forever
// return -2, parse error
func ExtractCacheLifetime(cacheLifeStr string) time.Duration {
	cacheLifeStr = strings.TrimSpace(cacheLifeStr)
	cacheLifeStr = strings.ToLower(cacheLifeStr)
	if "" == cacheLifeStr {
		return time.Duration(-1)
	}

	// check pattern
	cacheLifeAllReg := regexp.MustCompile(CACHE_LIFETIME_ALL_REG)
	if !cacheLifeAllReg.MatchString(cacheLifeStr) {
		log.Printf("[ERROR] failed to match cache lifetime string %s with pattern %s", cacheLifeStr, CACHE_LIFETIME_ALL_REG)
		return time.Duration(-2)
	}

	cacheLifeReg := regexp.MustCompile(CACHE_LIFETIME_REG)
	match := cacheLifeReg.FindAllStringSubmatch(cacheLifeStr, -1)
	if nil == match {
		log.Printf("[ERROR] failed to match cache lifetime string %s with pattern %s", cacheLifeStr, CACHE_LIFETIME_REG)
		return time.Duration(-2)
	}

	cacheTotalLife := time.Duration(0)
	for matInd, subMat := range match {
		if 3 != len(subMat) {
			log.Printf("[ERROR] len(submatch) != 3, cache lifetime string is %s and pattern is %s", cacheLifeStr, CACHE_LIFETIME_REG)
			return time.Duration(-2)
		}
		// subMat: [2d, 2, d]
		timeStr := subMat[1]
		unitStr := subMat[2]
		// check duplicate unit: second, minitue, hour or day
		for i := matInd - 1; i >= 0; i-- {
			if match[i][2] == unitStr {
				log.Printf("[ERROR] duplicate unit found in %s, %s and %s", cacheLifeStr, match[i][0], subMat[0])
				return time.Duration(-2)
			}
		}
		// parse time
		cacheLife, err := strconv.Atoi(timeStr)
		if nil != err {
			log.Printf("[ERROR] failed to parse cache time %s: %s", subMat[0], err)
			return time.Duration(-2)
		}

		timeUnit := time.Second
		switch unitStr {
		case "s":
			timeUnit = time.Second
		case "m":
			timeUnit = time.Minute
		case "h":
			timeUnit = time.Hour
		case "d":
			timeUnit = time.Hour * 24
		}
		cacheTotalLife += time.Duration(cacheLife) * timeUnit
	}

	return cacheTotalLife
}
