# gofeed

Gofeed was inspired by feed43.com. It is disigned to extract full-text feeds from websites which do not provide any.

This simple program was written when I started to learn golang. So I tried to reinvent everything I need, including a simple crawler which took good use of cache and a very simple rss2.0 feed generator.

## Features

* Like feed43.com, you can extract feed titles, feed links and the feed descriptions with the following predefined patterns. Note that all these patterns are **lazy** and perform **leftmost** match, which means they will match as few characters as possible.
    *  {any}: match any character including newline
    *  {title}: title of feed entry, matched against the Feed.URL page
    *  {link}: hyper link of feed entry, matched against the Feed.URL page
    *  {description}: full-text description of feed entry, matched against the corresponding {link} page
    *  custom regular expressions: **MUST NOT** contain any of the predefined patterns. And the syntax documentation can be found [here](https://code.google.com/p/re2/wiki/Syntax).
* The crawler knows well when to request new html and when to use local html cache. This will save a lot of bandwidth.
 
## Things need to be improved

*  Little documentation.
*  No enough tests(use coveralls.io coverage service).
*  Maybe more...

## More functions on the todo list

1. <del>Cache old requests: use sqlite to cache downloaded web pages and save their lastmod time.</del>
2. <del>Download html files for each feed target defined in the configuration in separate goroutines. </del>
3. <del>Add debug mode, which will print more debug infomation</del>
4. Add alternative methods to extract feed title, link and description from html
    1. xpath

## Install

Firstly, make sure you have set the `GOPATH` environment variable properly. Then, you should install the sqlite driver `go-sqlite3`.

    go get github.com/mattn/go-sqlite3

Now install gofeed.

    go get github.com/mawenbao/gofeed

## Configuration example

See `example_config.json` and `example_config2.json`.

*  CacheDB: (string) path of html cache database(sqlite3), can be absolute or be relative to the current directory.
*  Targets: array of feed targets, each of which runs in a separate goroutine
    *  Request.Interval: (integer) time to wait before sending a http request to the target.
    *  Feed.URL: (array of strings) array of urls, used to define urls of the target's index pages.
    *  Feed.IndexPattern: (array of strings) array of index patterns, used to extract entry link and entry title from the index page.
    *  Feed.ContentPattern: (array of strings) array of content patterns, used to extract entry description from the entry's content page(identified by entry's link).

And you should note that

1. There should be as many Feed.URL as Feed.IndexPattern. If array length of the two does not match, there should be only one Feed.IndexPattern, which means all the Feed.URL will use the same Feed.IndexPattern. Otherwise, an configuration parse error will return. 

    And the same goes for Feed.ContentPattern.

## Command line options

    Usage ./gofeed [-v][-d][-c cpu_number] json_config_file

    Flags:
    -c=2: number of cpus used to run simultaneously
    -v=false: be verbose
    -d=false: debug mode

*  -c: number of cpus, default value is the actual number of your machine's cpus.
*  -v: print more infomation.
*  -d: print even more information than `-v` option, should be useful when debugging your index or content patterns.

## License

BSD license, see LICENSE.txt for more details.

