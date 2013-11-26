package main

import(
    "strings"
    "log"
    "net/http"
)

// return later time
func GetLaterTimeStr(a, b string) (result string, err error) {
    timeA, err := http.ParseTime(a)
    if nil != err {
        log.Printf("failed to parse string %s as http time", a)
        return
    }
    timeB, err := http.ParseTime(b)
    if nil != err {
        log.Printf("failed to parse string %s as http time", b)
        return
    }

    if timeA.After(timeB) {
        result = a
    } else {
        result = b
    }

    return
}

// normalize url
func NormalizeURLStr(rawString string) string {
    if !strings.HasPrefix(rawString, HTTP_SCHEME) && !strings.HasPrefix(rawString, HTTPS_SCHEME) {
        rawString = HTTP_SCHEME + rawString
    }
    return rawString
}

