package main

import(
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
    "log"
    "os"
    "errors"
    "fmt"
)

// create the cache db with sqlite3 driver
func CreateDbScheme(dbPath string) (err error) {
    _, err = os.Stat(dbPath)
    if nil == err {
        // db file already exists
        log.Printf("db %s already exists", dbPath)
        return errors.New("db " + dbPath + " already exists")
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to create db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlCreateTables := fmt.Sprintf(`
    CREATE TABLE %s (
        id INTEGER PRIMARY KEY AUTOINCREMENT, 
        url TEXT NOT NULL UNIQUE,
        cache_control TEXT,
        lastmod TEXT,
        etag TEXT,
        expires TEXT,
        html BLOB
    );
    `, DB_HTML_CACHE_TABLE)

    _, err = db.Exec(sqlCreateTables)
    if nil != err {
        log.Printf("failed to create tables in db %s, sql is %s, error is %s", dbPath, sqlCreateTables, err)
        return
    }

    return
}

func PutHtmlCache(dbPath string, cache []HtmlCache) (err error) {
    _, err = os.Stat(dbPath)
    if nil != err {
        // db file not exists
        log.Printf("db %s not exists", dbPath)
        return
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlInsertHtml := fmt.Sprintf(`
    INSERT INTO %s (url, cache_control, lastmod, etag, expires, html) VALUES (?, ?, ?, ?, ?, ?);
    `, DB_HTML_CACHE_TABLE)

    trans, err := db.Begin()
    if nil != err {
        log.Printf("failed to start a new transaction for db %s: %s", dbPath, err)
        return
    }

    statmt, err := trans.Prepare(sqlInsertHtml)
    if nil != err {
        log.Printf("failed to prepare a new statement for db %s, sql %s: %s", dbPath, sqlInsertHtml, err)
        return
    }
    defer statmt.Close()

    urls := ""
    for _, c := range cache {
        _, err = statmt.Exec(c.URL, c.CacheControl, c.LastModified, c.Etag, c.Expires, c.Html)
        urls += c.URL + " "
        if nil != err {
            log.Printf("failed to insert html cache: %s", err)
            return
        }
    }

    err = trans.Commit()
    if nil != err {
        log.Printf("failed to save urls %s: %s", urls, err)
        return
    }

    if *gVerbose {
        log.Printf("successully saved cache for %s", urls)
    }
    return
}

func GetHtmlCacheByURL(dbPath, url string) (cache HtmlCache, err error) {
    _, err = os.Stat(dbPath)
    if nil != err {
        // db file not exists
        log.Printf("db %s not exists", dbPath)
        return
    }

    db, err := sql.Open(DB_DRIVER, dbPath)
    if nil != err {
        log.Printf("failed to open db %s: %s", dbPath, err)
        return
    }
    defer db.Close()

    sqlSelect := fmt.Sprintf("SELECT url, cache_control, lastmod, etag, expires, html FROM %s WHERE url = ?", DB_HTML_CACHE_TABLE)

    statmt, err := db.Prepare(sqlSelect)
    if nil != err {
        log.Printf("failed to prepare statment %s for db %s: %s", sqlSelect, dbPath, err)
        return
    }
    defer statmt.Close()

    err = statmt.QueryRow(url).Scan(
        &cache.URL,
        &cache.CacheControl,
        &cache.LastModified,
        &cache.Etag,
        &cache.Expires,
        &cache.Html)
    if nil != err {
        //log.Printf("failed to select html cache with url %s: %s", url, err)
        return
    }

    return
}

