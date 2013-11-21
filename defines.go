package main

type Target struct {
    URL string
    IndexPattern string
    ContentPattern string
    Path string
}

type TargetSlice struct {
    Targets []Target
}
