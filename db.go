package main

import (
	"bytes"
	"compress/gzip"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// create the cache db with sqlite3 driver
func CreateDBScheme(dbPath string) (err error) {
	_, err = os.Stat(dbPath)
	if nil == err {
		// db file already exists
		log.Printf("[ERROR] db %s already exists", dbPath)
		return errors.New("db " + dbPath + " already exists")
	}

	db, err := sql.Open(DB_DRIVER, dbPath)
	if nil != err {
		log.Printf("[ERROR] failed to create db %s: %s", dbPath, err)
		return
	}
	defer db.Close()

	sqlCreateTables := fmt.Sprintf(`
    CREATE TABLE %s (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        url TEXT NOT NULL UNIQUE,
        date TEXT NOT NULL,
        cache_control TEXT,
        lastmod TEXT,
        etag TEXT,
        expires TEXT,
        html BLOB
    );
    `, DB_HTML_CACHE_TABLE)

	_, err = db.Exec(sqlCreateTables)
	if nil != err {
		log.Printf("[ERROR] failed to create tables in db %s, sql is %s, error is %s", dbPath, sqlCreateTables, err)
		return
	}

	sqlCreateIndex := fmt.Sprintf(`
    CREATE UNIQUE INDEX IF NOT EXISTS html_cache_url_index ON %s (url);
    `, DB_HTML_CACHE_TABLE)

	_, err = db.Exec(sqlCreateIndex)
	if nil != err {
		log.Printf("[ERROR] failed to crate url index in db %s, sql is %s, error is %s", dbPath, sqlCreateIndex, err)
		return
	}

	return
}

func ExecQuerySQL(dbPath string, expectSize int, sqlStr string, args ...interface{}) (caches []*HtmlCache, err error) {
	_, err = os.Stat(dbPath)
	if nil != err {
		// db file not exists
		log.Printf("[ERROR] db %s not exists", dbPath)
		return
	}

	db, err := sql.Open(DB_DRIVER, dbPath)
	if nil != err {
		log.Printf("[ERROR] failed to open db %s: %s", dbPath, err)
		return
	}
	defer db.Close()

	statmt, err := db.Prepare(sqlStr)
	if nil != err {
		log.Printf("[ERROR] failed to prepare statment %s for db %s: %s", sqlStr, dbPath, err)
		return
	}
	defer statmt.Close()

	rows, err := statmt.Query(args...)
	if nil != err {
		log.Printf("[ERROR] failed to query with statment %s, %s", sqlStr, err)
		return
	}
	defer rows.Close()

	if expectSize > 0 {
		caches = make([]*HtmlCache, expectSize)
	}
	rowInd := 0
	for rows.Next() {
		c := new(HtmlCache)
		var urlStr, lastmod, expires, dateStr string
		err = rows.Scan(
			&urlStr,
			&dateStr,
			&c.CacheControl,
			&lastmod,
			&c.Etag,
			&expires,
			&c.Html)
		if nil != err {
			log.Printf("[ERROR] failed to scan data from result row: %s", err)
			return
		}

		// decompress html data
		if 0 != *gGzipCompressLevel {
			buff := bytes.NewBuffer(c.Html)
			gzipR, err := gzip.NewReader(buff)
			if nil != err {
				if *gDebug {
					log.Printf("[WARN] failed to decompress html data for %s: %s", urlStr, err)
				}
			} else {
				c.Html, err = ioutil.ReadAll(gzipR)
			}
		}

		if c.URL, err = url.Parse(urlStr); nil != err {
			log.Printf("[ERROR] failed to parse url from rawurl string %s: %s", urlStr, err)
		}
		if "" != lastmod {
			c.LastModified = new(time.Time)
			if *c.LastModified, err = http.ParseTime(lastmod); nil != err {
				log.Printf("[ERROR] failed to parse lastmod time string %s: %s", lastmod, err)
			}
		}
		if "" != expires {
			c.Expires = new(time.Time)
			if *c.Expires, err = http.ParseTime(expires); nil != err {
				log.Printf("[ERROR] failed to parse expires time string %s: %s", expires, err)
			}
		}
		if "" != dateStr {
			c.Date = new(time.Time)
			if *c.Date, err = http.ParseTime(dateStr); nil != err {
				log.Printf("[ERROR] failed to parse cache date %s: %s", dateStr, err)
			}
		}

		caches = append(caches[:rowInd], c)
		rowInd += 1
	}

	// no result is also an error
	if 0 == rowInd {
		err = DBNoRecordError{}
		return
	}

	return
}

func ExecInsertUpdateSQL(caches []*HtmlCache, dbPath string, sqlStr string) (err error) {
	_, err = os.Stat(dbPath)
	if nil != err {
		// db file not exists
		log.Printf("[ERROR] db %s not exists", dbPath)
		return
	}

	db, err := sql.Open(DB_DRIVER, dbPath)
	if nil != err {
		log.Printf("[ERROR] failed to open db %s: %s", dbPath, err)
		return
	}
	defer db.Close()

	trans, err := db.Begin()
	if nil != err {
		log.Printf("[ERROR] failed to start a new transaction for db %s: %s", dbPath, err)
		return
	}

	statmt, err := trans.Prepare(sqlStr)
	if nil != err {
		log.Printf("[ERROR] failed to prepare a new statement for db %s, sql %s: %s", dbPath, sqlStr, err)
		return
	}
	defer statmt.Close()

	urls := ""
	for _, c := range caches {
		if nil == c {
			log.Printf("[ERROR] cache is nil, ignore this one")
			continue
		}
		if nil == c.Date {
			log.Printf("[ERROR] cache date is nil, will not save this sucker in cache db")
			continue
		}

		var htmlBuff bytes.Buffer
		compressed := false

		if 0 != *gGzipCompressLevel {
			// compress html data
			gzipW, err := gzip.NewWriterLevel(&htmlBuff, *gGzipCompressLevel)
			if nil != err {
				log.Printf("[ERROR] failed to create gzip writer: %s", err)
				continue
			}
			_, err = gzipW.Write(c.Html)
			if nil != err {
				// on write error, cache html is saved uncompressed
				log.Printf("[ERROR] gzip failed to compress html data: %s, will not compress the html data", err)
			}
			gzipW.Close()
			compressed = true
		}
		htmlData := c.Html
		if compressed {
			htmlData = htmlBuff.Bytes()
		}

		var lastmod, expires string
		if nil != c.LastModified {
			lastmod = c.LastModified.Format(http.TimeFormat)
		}
		if nil != c.Expires {
			expires = c.Expires.Format(http.TimeFormat)
		}
		_, err = statmt.Exec(c.URL.String(), c.Date.Format(http.TimeFormat), c.CacheControl, lastmod, c.Etag, expires, htmlData)
		urls += c.URL.String() + " "
		if nil != err {
			log.Printf("[ERROR] failed to exec insert/update sql %s: %s", sqlStr, err)
			return
		}
	}

	err = trans.Commit()
	if nil != err {
		log.Printf("[ERROR] failed to save urls %s: %s", urls, err)
		return
	}

	return
}

func GetHtmlCacheByURL(dbPath, urlStr string) (cache *HtmlCache, err error) {
	htmlCacheSlice, err := ExecQuerySQL(
		dbPath,
		1,
		fmt.Sprintf("SELECT url, date, cache_control, lastmod, etag, expires, html FROM %s WHERE url = ?", DB_HTML_CACHE_TABLE),
		urlStr)

	if nil != err {
		switch err.(type) {
		case DBNoRecordError:
			if *gVerbose {
				log.Printf("cache not found for %s", urlStr)
			}
		default:
			log.Printf("[ERROR] failed to get cache from db %s by url %s: %s", dbPath, urlStr, err)
		}
		return nil, err
	}

	return htmlCacheSlice[0], err
}

func PutHtmlCache(dbPath string, caches []*HtmlCache) (err error) {
	sqlInsertHtml := fmt.Sprintf(`
    INSERT INTO %s (url, date, cache_control, lastmod, etag, expires, html) VALUES (?, ?, ?, ?, ?, ?, ?);
    `, DB_HTML_CACHE_TABLE)

	err = ExecInsertUpdateSQL(caches, dbPath, sqlInsertHtml)
	if nil != err {
		log.Printf("[ERROR] failed to insert cache records to db %s: %s", dbPath, err)
		return
	}

	if *gVerbose {
		for _, c := range caches {
			log.Printf("successully saved cache for %s", c.URL.String())
		}
	}
	return
}

func UpdateHtmlCache(dbPath string, caches []*HtmlCache) (err error) {
	sqlUpdateHtml := ""
	for _, c := range caches {
		sqlUpdateHtml += fmt.Sprintf(`
        UPDATE %s SET url = ?, date = ?, cache_control = ?, lastmod = ?, etag = ?, expires = ?, html = ? WHERE url = '%s';
        `, DB_HTML_CACHE_TABLE, c.URL.String())
	}
	err = ExecInsertUpdateSQL(caches, dbPath, sqlUpdateHtml)
	if nil != err {
		log.Printf("[ERROR] failed to update cache records to db %s: %s", dbPath, err)
		return
	}

	if *gVerbose {
		for _, c := range caches {
			log.Printf("successully updated cache for %s", c.URL.String())
		}
	}
	return
}
