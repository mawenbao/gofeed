package main

import (
	"log"
	"regexp"
	"strings"
)

func PatternToRegex(pat string) string {
	r := strings.NewReplacer(
		PATTERN_ANY, PATTERN_ANY_REG,
		PATTERN_TITLE, PATTERN_TITLE_REG,
		PATTERN_LINK, PATTERN_LINK_REG,
		PATTERN_CONTENT, PATTERN_CONTENT_REG)

	return r.Replace(pat)
}

func CheckPatterns(tar *TargetConfig) bool {
	if nil == tar {
		log.Printf("[ERROR] invliad target, nil")
		return false
	}

	if (len(tar.URLs) != len(tar.IndexPatterns) && 1 != len(tar.IndexPatterns)) ||
		(len(tar.URLs) != len(tar.ContentPatterns) && 1 != len(tar.ContentPatterns)) {
		log.Printf("error parsing index/content patterns: len(URL) != len(IndexPattern|ContentPattern) || 1 != len(IndexPattern|ContentPattern")
		return false
	}

	for _, indexPat := range tar.IndexPatterns {
		// IndexPattern should contain both {title} and {link}
		if "" == indexPat {
			log.Print("[ERROR] index pattern is empty")
			return false
		}

		if 1 != strings.Count(indexPat, PATTERN_TITLE) || 1 != strings.Count(indexPat, PATTERN_LINK) {
			log.Printf("[ERROR] index pattern %s should contain 1 %s and 1 %s ", indexPat, PATTERN_TITLE, PATTERN_LINK)
			return false
		}
	}

	for _, contentPat := range tar.ContentPatterns {
		// ContentPattern should contain {content} and should not contain {title} nor {link}
		if "" == contentPat {
			log.Print("content pattern is empty")
			return false
		}

		if 1 != strings.Count(contentPat, PATTERN_CONTENT) {
			log.Printf("[ERROR] content pattern %s should contain 1 %s", contentPat, PATTERN_CONTENT)
			return false
		}

		if strings.Contains(contentPat, PATTERN_TITLE) || strings.Contains(contentPat, PATTERN_LINK) {
			log.Printf("[ERROR] %s should not contain %s or %s", contentPat, PATTERN_TITLE, PATTERN_LINK)
			return false
		}
	}

	return true
}

func CompileIndexContentPatterns(feedTar *FeedTarget, tar *TargetConfig) (err error) {
	feedTar.IndexRegs = make([]*regexp.Regexp, len(tar.IndexPatterns))
	feedTar.ContentRegs = make([]*regexp.Regexp, len(tar.ContentPatterns))

	// index pattern
	if 1 == len(tar.IndexPatterns) {
		feedTar.IndexRegs[0], err = regexp.Compile(PatternToRegex(tar.IndexPatterns[0]))
		if nil != err {
			log.Printf("[ERROR] error compiling index pattern %s", tar.IndexPatterns[0])
			return
		}
	} else {
		for j := 0; j < len(tar.IndexPatterns); j++ {
			feedTar.IndexRegs[j], err = regexp.Compile(PatternToRegex(tar.IndexPatterns[j]))
			if nil != err {
				log.Printf("[ERROR] error compiling index pattern %s", tar.IndexPatterns[j])
				return
			}
		}
	}

	// content pattern
	if 1 == len(tar.ContentPatterns) {
		feedTar.ContentRegs[0], err = regexp.Compile(PatternToRegex(tar.ContentPatterns[0]))
		if nil != err {
			log.Printf("[ERROR] error compiling content pattern %s", tar.ContentPatterns[0])
			return
		}
	} else {
		for j := 0; j < len(tar.ContentPatterns); j++ {
			feedTar.ContentRegs[j], err = regexp.Compile(PatternToRegex(tar.ContentPatterns[j]))
			if nil != err {
				log.Printf("[ERROR] error compiling content pattern %s", tar.ContentPatterns[j])
				return
			}
		}
	}

	return
}
