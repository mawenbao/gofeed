# gofeed

Gofeed was inspired by feed43.com.

**NOT FINISHED YET**

gofeed parses some web pages into full-text feed. This simple program is written when I start to learn golang. So it may not be ready to use currently.

## Dependency

*  go-sqlite3

        go get github.com/mattn/go-sqlite3

## Install

    go get github.com/mawenbao/gofeed

## Configuration Example
See example_config.json.

## Supported Patterns
Note that all the regular expressions are lazy.

*  {any}: match any character including newline
*  {title}: title of feed entry, title is matched against the Feed.URL page
*  {link}: hyper link of feed entry, link is matched against the Feed.URL page
*  {content}: full-text content of feed entry, content is matched against the corresponding {link} page
 
## TODO

1. Cache old requests: use sqlite to cache downloaded web pages and save their lastmod time.
2. Add alternative methods to extract feed title, link and content from html
    1. xpath
3. Better readme file

## LICENSE

BSD license, see LICENSE.txt for more details.

