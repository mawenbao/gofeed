# gofeed

Gofeed was inspired by feed43.com. It is disigned to extract full-text feeds from websites which only provide partial feeds or provide no feeds at all.

This simple program was written when I started to learn golang. So I tried to reinvent everything I need, including a simple crawler which took good use of cache and a very simple rss2.0 feed generator.

## Releases
See [http://dl.atime.me/gofeed](http://dl.atime.me/gofeed).

## Features

* Like feed43.com, you can extract feed titles, feed links and the feed descriptions from web pages or partial feeds with some predefined patterns and even your custom regular expressions. 
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

### Json configuration
See `example_config.json` and `example_config2.json`.

*  CacheDB: (string) path of html cache database(sqlite3), can be absolute or be relative to the current directory.
*  CacheLifetime: (string) cache will be removed from cache database after CacheLifetime, "", "5d", "3h", "10m", "1s" and "5d3h10m" are acceptable where "" means cache lives forever in the cache database.
*  Targets: array of feed targets, each of which runs in a separate goroutine
    *  Request.Interval: (integer) time to wait before sending a http request to the target.
    *  Feed.Path: (string) output path of the rss2 feed file, can be relative or absolute.
    *  Feed.Title: (string) title of the rss2 feed channel. If not defined, feed title will be the filename of Feed.Path.
    *  Feed.Description: (string) description of the rss2 feed channel. In not defined, feed description will be empty.
    *  Feed.URL: (array of strings) array of urls, used to define urls of the target's index pages. Note that this url can be html or xml or anything that you can extract feed entry titles and links with regex patterns.
    *  Feed.IndexFilterPattern: (array of strings) array of index filter patterns, used to filter valid index html from the entire html.
    *  Feed.IndexPattern: (array of strings) array of index patterns, used to extract entry link and entry title from the filtered content by Feed.IndexFilterPattern.
    *  Feed.ContentFilterPattern: (array of strings) array of content patterns, used to extract valid content html from the entire html identified by {link}.
    *  Feed.ContentPattern: (array of strings) array of content patterns, used to extract entry description from the entry's filtered html content by Feed.ContentFilterPattern.
    *  Feed.PubDatePattern: (string) pattern of publish date, see pre-defined patterns. Used to extract publish date of an article from the string extracted from {pubdate} pattern.

And you should note that

1. There should be as many Feed.URL as Feed.IndexPattern. If array length of the two does not match, there should be only one Feed.IndexPattern or only one Feed.URL, which means all the Feed.URL will share the same Feed.IndexPattern or all the Feed.IndexPattern share the same Feed.URL. Otherwise, an configuration parse error will return. 

2. For Feed.ContentPattern, there should be as many Feed.URL as Feed.ContentPattern. If array length of the two does not match, there should be only one Feed.ContentPattern, which means all the Feed.URL will share the same Feed.ContentPattern. And the same goes for Feed.PubDateFormat.

3. Either Feed.IndexPattern or Feed.ContentPattern can contain the {pubdate} pattern, but not both.

### Pre-defined patterns
You can use the following predefined patterns in `Feed.IndexPattern` and `Feed.ContentPattern` of the json configuration. Note that all these patterns are **lazy** and perform **leftmost** match, which means they will match as few characters as possible.

*  {any}: match any character including newline
*  {title}: title of feed entry, matched against the Feed.URL page
*  {link}: hyper link of feed entry, matched against the Feed.URL page
*  {description}: full-text description of feed entry, matched against the corresponding {link} page
*  {pubdate}: publish date of feed entry
*  {filter}: filtered content, used in Feed.IndexFilterPattern or Feed.ContentFilterPattern

Date time format pattern, currently used for publish date string extraced from the {pubdate} pattern. Note that, unlike other pre-defined patterns, all these date related patterns are greedy.

*  {year}: must be an integer
*  {month}: must be an integer, Jan or January.
*  {day}: must be an integer
*  {hour}: must be an integer
*  {minute}: must be an integer
*  {second}: must be an integer

### Custom regular expressions
You can also write custom regex in `Feed.IndexPattern` and `Feed.ContentPattern`. Make sure there are no predefined patterns in your custom regular expressions. The regex syntax documentation can be found [here](https://code.google.com/p/re2/wiki/Syntax).

The custom regular expressions have not been tested properly. So I suggest just using the predefined patterns.

## Command line options

    Usage ./gofeed [-version][-v][-d][-c cpu_number][-l log_file][-k][-z compression_level] json_config_file

    Flags:
    -a=false: use cache if failed to download web page
    -c=2: number of cpus to run simultaneously
    -v=false: be verbose
    -d=false: debug mode
    -l="": path of the log file
    -k=false: keep feed entries which do not have any description
    -z=9: compression level when saving html cache with gzip in the cache database.
        0-9 acceptable where 0 means no compression
    -version=false: print gofeed version

*  -a: If failed to download the target url, try to use cache even it has expired.
*  -c: Number of cpus, default value is the actual number of your machine's cpus.
*  -v: Print more infomation.
*  -d: Print even more information than `-v` option, should be useful when debugging your index or content patterns.
*  -l: Append output in a log file
*  -z: Gofeed compresses html cache data with gzip by default. This option can set compression level of gzip, however, you can pass 0 to disable compression.
*  -version: Print gofeed version

## License

BSD license, see LICENSE.txt for more details.

