# gofeed

Gofeed was inspired by feed43.com. It is disigned to extract full-text feeds from websites which do not provide any.

This simple program was written when I started to learn golang. So I tried to reinvent everything I need, including a simple crawler which took good use of cache and a very simple rss2.0 feed generator.

## Features

* Like feed43.com, you can extract feed titles, feed links and the feed descriptions with the following predefined patterns. Note that all these patterns are **lazy** and perform **leftmost** match, which means they will match as few characters as possible.
    *  {any}: match any character including newline
    *  {title}: title of feed entry, matched against the Feed.URL page
    *  {link}: hyper link of feed entry, matched against the Feed.URL page
    *  {description}: full-text description of feed entry, matched against the corresponding {link} page
    *  custom regular expressions: *MUST NOT* contain any of the predefined patterns
* The crawler knows well when to request new html and when to use local html cache. This will save a lot of bandwidth.
 
## Things need to be improved

*  Little documentation.
*  No enough tests(use coveralls.io coverage service)
*  Maybe more...

## More functions on the todo list

1. <del>Cache old requests: use sqlite to cache downloaded web pages and save their lastmod time.</del>
2. Download html files for each feed target defined in the configuration in separate goroutines. 
3. Add debug mode, which prints more infomation
4. Add alternative methods to extract feed title, link and description from html
    1. xpath

## Install

Firstly, you should install the sqlite driver `go-sqlite3`.

    go get github.com/mattn/go-sqlite3

Then install gofeed.

    go get github.com/mawenbao/gofeed

## Configuration example

See `example_config.json`.

## License

BSD license, see LICENSE.txt for more details.

