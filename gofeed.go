package main

import (
    "os"
    "fmt"
    "flag"
)

var (
    gVerbose = flag.Bool("v", false, "be verbose or not")
    gFulltext = flag.Bool("f", false, "generate full-text feeds or not")
)

func showUsage() {
    fmt.Printf("Usage %s [-v][-f] URLS ...\n\n", os.Args[0])
    fmt.Printf("Flags:\n\n")
    flag.PrintDefaults()
}

func main() {
    flag.Usage = showUsage
    flag.Parse()

    urls := flag.Args()
    if 0 == len(urls) {
        flag.Usage()
        return
    }

    fmt.Println("urls", urls)
}

