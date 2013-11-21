package main

const(
    SCHEME_SUFFIX = "://"
    HTTP_SCHEME = "http"
)

type Target struct {
    URL string
    IndexPattern string
    ContentPattern string
    Path string
}

type TargetSlice struct {
    Targets []Target
}

