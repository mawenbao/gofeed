package main

import(
    "os"
    "fmt"
    "flag"
    "log"
)

func init() {
    log.SetFlags(log.Lshortfile)
}

var (
    gVerbose = flag.Bool("v", true, "be verbose")
    gFulltext = flag.Bool("f", false, "generate full-text feeds")
)

func showUsage() {
    fmt.Printf("Usage %s [-v][-f] json_config_file\n\n", os.Args[0])
    fmt.Printf("Flags:\n\n")
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

    /*
    _, err := ParseJsonConfig(args[0])
    fmt.Println(err)

    data, _ := Crawl("blog.atime.me/agreement.html")
    fmt.Println(string(data))
    */
}

